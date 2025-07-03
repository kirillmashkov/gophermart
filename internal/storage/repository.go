package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kirillmashkov/gophermart/internal/model"
	"go.uber.org/zap"
)

type RepositoryUser struct {
	db  *Database
	log *zap.Logger
}

const timeoutOperationDB = 1 * time.Second

func NewRepositoryUser(db *Database, log *zap.Logger) *RepositoryUser {
	return &RepositoryUser{db: db, log: log}
}

func (r *RepositoryUser) RegisterUser(ctx context.Context, login string, password string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutOperationDB)
	defer cancel()

	tx, err := r.db.Dbpool.Begin(ctx)
	if err != nil {
		r.log.Error("Error open tran", zap.Error(err))
		return "", err
	}
	defer func() {
		if err == nil {
			if errCommit := tx.Commit(ctx); errCommit != nil {
				r.log.Error("Error commit tran", zap.Error(err))
			}
		} else {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				r.log.Error("Error rollback tx", zap.Error(errRollback))
			}
		}
	}()

	userID := uuid.NewString()
	_, err = tx.Exec(ctx, "insert into profile (id, login, password, balance, withdrawn) values ($1, $2, $3, 0, 0)", userID, login, password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return "", model.ErrDuplicateLogin
			}
		}
		return "", err
	}

	return userID, nil
}

func (r *RepositoryUser) GetUserID(ctx context.Context, login string, password string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutOperationDB)
	defer cancel()

	var id string
	err := r.db.Dbpool.QueryRow(ctx, "select id from profile where login = $1 and password = $2", login, password).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		r.log.Error("No user found or password is incorrect")
		return false, "", nil
	}

	if err != nil {
		r.log.Error("Error when try to login user")
		return false, "", nil
	}

	return true, id, nil
}

func (r *RepositoryUser) GetOrderByOrderNum(ctx context.Context, orderNum int64) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutOperationDB)
	defer cancel()

	var profileID string
	err := r.db.Dbpool.QueryRow(ctx, "select profile_id from orders where order_num = $1", orderNum).Scan(&profileID)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, "", nil
	}

	if err != nil {
		r.log.Error("Error when GetOrderByOrderNum", zap.Int64("orderNum", orderNum))
		return false, "", nil
	}

	return true, profileID, nil
}

func (r *RepositoryUser) CreateBalanceOrder(orderNum int64, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutOperationDB)
	defer cancel()

	tx, err := r.db.Dbpool.Begin(ctx)
	if err != nil {
		r.log.Error("Error open tran", zap.Error(err))
		return "", err
	}

	defer func() {
		if err == nil {
			if errCommit := tx.Commit(ctx); errCommit != nil {
				r.log.Error("Error commit tran", zap.Error(err))
			}
		} else {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				r.log.Error("Error rollback tx", zap.Error(errRollback))
			}
		}
	}()

	orderID := uuid.NewString()
	_, err = tx.Exec(ctx, "insert into orders (id, profile_id, order_num, status, uploaded_at, type_order) values ($1, $2, $3, $4, $5, $6)", orderID, userID, orderNum, "NEW", time.Now(), "BALANCE")
	if err != nil {
		r.log.Error("Error insert into order", zap.Int64("orderNum", orderNum), zap.String("userID", userID), zap.Error(err))
		return "", err
	}

	return orderID, nil
}

func (r *RepositoryUser) CreateWithdrawnOrder(ctx context.Context, orderNum int64, userID string, sum float32) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutOperationDB)
	defer cancel()

	tx, err := r.db.Dbpool.Begin(ctx)
	if err != nil {
		r.log.Error("Error open tran", zap.Error(err))
		return err
	}

	defer func() {
		if err == nil {
			if errCommit := tx.Commit(ctx); errCommit != nil {
				r.log.Error("Error commit tran", zap.Error(err))
			}
		} else {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				r.log.Error("Error rollback tx", zap.Error(errRollback))
			}
		}
	}()

	orderID := uuid.NewString()
	_, err = tx.Exec(ctx, "insert into orders (id, profile_id, order_num, status, uploaded_at, type_order, sum) values ($1, $2, $3, $4, $5, $6, $7)", orderID, userID, orderNum, "PROCESSED", time.Now(), "WITHDRAW", sum)
	if err != nil {
		r.log.Error("Error insert into order", zap.Int64("orderNum", orderNum), zap.String("userID", userID), zap.Error(err))
		return err
	}

	_, err = tx.Exec(ctx, "update profile set balance=balance-$1, withdrawn=withdrawn+$1 where id=$2 and balance-$1>=0 returning id", sum, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			r.log.Error("No balance for user", zap.String("UserID", userID), zap.Int64("OrderNum", orderNum))
			return model.ErrNoBalance
		}

		r.log.Error("Can't process withdrawn order", zap.String("UserID", userID), zap.Int64("orderNum", orderNum), zap.Error(err))
	}

	return nil
}

func (r *RepositoryUser) UpdateOrder(id string, status string, accrual float32, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutOperationDB)
	defer cancel()

	tx, err := r.db.Dbpool.Begin(ctx)
	if err != nil {
		r.log.Error("Error open tran", zap.Error(err))
		return err
	}

	defer func() {
		if err == nil {
			if errCommit := tx.Commit(ctx); errCommit != nil {
				r.log.Error("Error commit tran", zap.Error(err))
			}
		} else {
			if errRollback := tx.Rollback(ctx); errRollback != nil {
				r.log.Error("Error rollback tx", zap.Error(errRollback))
			}
		}
	}()

	_, err = tx.Exec(ctx, "update orders set status = $1, sum = $2 where id = $3", status, accrual, id)
	if err != nil {
		r.log.Error("Error update order", zap.String("id", id), zap.Error(err))
		return err
	}

	_, err = tx.Exec(ctx, "update profile set balance = balance + $1 where id = $2", accrual, userID)
	if err != nil {
		r.log.Error("Error update balance in profile", zap.String("userID", userID), zap.Error(err))
		return err
	}

	return nil
}

func (r *RepositoryUser) GetOrders(ctx context.Context, userID string, typeOrder string) ([]model.OrdersDB, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutOperationDB)
	defer cancel()

	rows, err := r.db.Dbpool.Query(ctx, "select order_num, status, sum, uploaded_at from orders where profile_id = $1 and type_order = $2 order by uploaded_at desc", userID, typeOrder)
	if err != nil {
		r.log.Error("Error get orders", zap.String("UserID", userID), zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	res, err := pgx.CollectRows(rows, pgx.RowToStructByPos[model.OrdersDB])
	if err != nil {
		r.log.Error("Error collect rows", zap.String("UserID", userID), zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (r *RepositoryUser) GetBalance(ctx context.Context, userID string) (model.BalanceResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, timeoutOperationDB)
	defer cancel()

	balance := model.BalanceResponse{}
	err := r.db.Dbpool.QueryRow(ctx, "select balance, withdrawn from profile where id = $1", userID).Scan(&balance.Balance, &balance.Withdrawn)

	if err != nil {
		r.log.Error("Error get balance", zap.String("UserID", userID))
		return model.BalanceResponse{}, err
	}

	return balance, nil
}
