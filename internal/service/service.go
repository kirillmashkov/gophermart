package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/kirillmashkov/gophermart/internal/config"
	"github.com/kirillmashkov/gophermart/internal/model"
	"github.com/kirillmashkov/gophermart/internal/storage"
	"github.com/kirillmashkov/gophermart/internal/util"
	"go.uber.org/zap"
)

type ServiceUser struct {
	log            *zap.Logger
	repositoryuser *storage.RepositoryUser
	securityUtil   *util.SecurityUtil
	cfg            *config.ServerConfig
}

func NewServiceUser(log *zap.Logger, ru *storage.RepositoryUser, su *util.SecurityUtil, cfg *config.ServerConfig) *ServiceUser {
	return &ServiceUser{log: log, repositoryuser: ru, securityUtil: su, cfg: cfg}
}

func (s *ServiceUser) RegisterUser(ctx context.Context, login string, password string) (string, error) {
	s.log.Info("Register new profile", zap.String("login", login), zap.String("password", password))
	userID, err := s.repositoryuser.RegisterUser(ctx, login, password)
	if err != nil {
		s.log.Error("Can't register new user", zap.Error(err))
		return "", err
	}

	token, err := s.securityUtil.BuildJWTString(userID)
	if err != nil {
		s.log.Error("Can't generate token", zap.Error(err))
		return "", err
	}

	return token, nil
}

func (s *ServiceUser) LoginUser(ctx context.Context, login string, password string) (bool, string, error) {
	checkUser, userID, err := s.repositoryuser.GetUserID(ctx, login, password)

	if err != nil {
		return false, "", err
	}

	var token = ""
	if checkUser {
		token, err = s.securityUtil.BuildJWTString(userID)
		if err != nil {
			s.log.Error("Error build jwt")
			return false, "", err
		}
	}

	return checkUser, token, nil
}

func (s *ServiceUser) Order(ctx context.Context, orderNum int64, userID string) error {
	validOrderNum := s.validLuhn(orderNum)

	if !validOrderNum {
		s.log.Error("Error order num", zap.Int64("orderNum", orderNum), zap.String("userID", userID))
		return model.ErrWrongOrderNumber
	}

	orderExist, profileID, err := s.repositoryuser.GetOrderByOrderNum(ctx, orderNum)
	if err != nil {
		return err
	}

	if orderExist {
		if profileID == userID {
			return model.ErrOrderAlreadyExistSameUser
		} else {
			return model.ErrOrderAlreadyExistOtherUser
		}
	}

	orderToCreate := model.OrderToCreate{OrderNum: orderNum, UserID: userID}
	model.OrderNumChanToCreate <- orderToCreate

	return nil
}

func (s *ServiceUser) RequestOrderAccrual() {
	for o := range model.RequestToAccrualChan {
		resp, err := http.Get(s.cfg.AccrualAddress + "/api/orders/" + strconv.FormatInt(o.OrderNum, 10))
		if err != nil {
			s.log.Error("Can't get request order from accrual", zap.Int64("orderNum", o.OrderNum), zap.Error(err))
			go s.retryRequestAccrual(o)
			continue
		}

		if resp.StatusCode != 200 {
			s.log.Error("Can't get order from accrual", zap.Int64("orderNum", o.OrderNum), zap.Int("status", resp.StatusCode), zap.Error(err))
			go s.retryRequestAccrual(o)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			s.log.Error("Can't get body from response accrual", zap.Int64("orderNum", o.OrderNum), zap.Error(err))
			go s.retryRequestAccrual(o)
			continue
		}

		accrualReponse := &model.AccrualResponse{}

		err = json.Unmarshal(body, accrualReponse)
		if err != nil {
			s.log.Error("Can't parse reponse from accrual", zap.Binary("Body", body), zap.Int64("orderNum", o.OrderNum), zap.Error(err))
			go s.retryRequestAccrual(o)
			continue
		}

		resp.Body.Close()

		err = s.repositoryuser.UpdateOrder(o.OrderID, accrualReponse.Status, accrualReponse.Accrual, o.UserID)
		if err != nil {
			go s.retryRequestAccrual(o)
			continue
		}
	}
}

func (s *ServiceUser) CreateBalanceOrder() {
	for o := range model.OrderNumChanToCreate {
		orderID, err := s.repositoryuser.CreateBalanceOrder(o.OrderNum, o.UserID)
		if err != nil {
			s.log.Error("Can't create order in db", zap.Int64("OrderNum", o.OrderNum))
			continue
		}

		orderToRequestAccrual := model.OrderToRequestAccrual{OrderNum: o.OrderNum, OrderID: orderID, UserID: o.UserID}
		model.RequestToAccrualChan <- orderToRequestAccrual
	}
}

func (s *ServiceUser) retryRequestAccrual(o model.OrderToRequestAccrual) {
	model.RequestToAccrualChan <- o
}

func (s *ServiceUser) validLuhn(orderNum int64) bool {
	return (orderNum%10+s.checksum(orderNum/10))%10 == 0
}

func (s *ServiceUser) checksum(number int64) int64 {
	var luhn int64

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}
	return luhn % 10
}

func (s *ServiceUser) GetBalanceOrders(ctx context.Context, userID string) ([]model.OrdersBalanceReposponse, error) {
	ordersDB, err := s.repositoryuser.GetOrders(ctx, userID, "BALANCE")
	if err != nil {
		return nil, err
	}

	ordersBalance := make([]model.OrdersBalanceReposponse, len(ordersDB))
	for i, val := range ordersDB {
		ordersBalance[i].Accrual = val.Sum
		ordersBalance[i].Number = val.Number
		ordersBalance[i].Status = val.Status
		ordersBalance[i].UploadedAt = val.UploadedAt
	}
	return ordersBalance, nil
}

func (s *ServiceUser) GetWithdrawalOrders(ctx context.Context, userID string) ([]model.OrderWithdrawlResponse, error) {
	ordersDB, err := s.repositoryuser.GetOrders(ctx, userID, "WITHDRAW")
	if err != nil {
		return nil, err
	}

	ordersWithdrawn := make([]model.OrderWithdrawlResponse, len(ordersDB))
	for i, val := range ordersDB {
		ordersWithdrawn[i].Order = val.Number
		ordersWithdrawn[i].Sum = val.Sum
		ordersWithdrawn[i].ProcessedAt = val.UploadedAt
	}

	return ordersWithdrawn, nil
}

func (s *ServiceUser) GetBalance(ctx context.Context, userID string) (model.BalanceResponse, error) {
	return s.repositoryuser.GetBalance(ctx, userID)
}

func (s *ServiceUser) CreateWithdrawnOrder(ctx context.Context, userID string, orderNum int64, sum int) error {
	validOrderNum := s.validLuhn(orderNum)

	if !validOrderNum {
		s.log.Error("Error order num", zap.Int64("orderNum", orderNum), zap.String("userID", userID))
		return model.ErrWrongOrderNumber
	}

	return s.repositoryuser.CreateWithdrawnOrder(ctx, orderNum, userID, sum)
}
