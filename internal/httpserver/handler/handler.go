package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/kirillmashkov/gophermart/internal/app"
	"github.com/kirillmashkov/gophermart/internal/httpserver/middleware/security"
	"github.com/kirillmashkov/gophermart/internal/model"
	"go.uber.org/zap"
)

func RegisterUser(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	var request model.LoginPassRequest
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		app.Log.Debug("cannot parse request JSON body", zap.Error(err))
		http.Error(res, "cannot parse request JSON body", http.StatusBadRequest)
		return
	}

	token, err := app.ServiceUser.RegisterUser(req.Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, model.ErrDuplicateLogin) {
			res.WriteHeader(http.StatusConflict)
			return
		}
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.Header().Set("Authorization", "Bearer "+token)
	res.WriteHeader(http.StatusOK)
}

func LoginUser(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	var request model.LoginPassRequest
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		app.Log.Debug("cannot parse request JSON body", zap.Error(err))
		http.Error(res, "cannot parse request JSON body", http.StatusBadRequest)
		return
	}

	checkUser, token, err := app.ServiceUser.LoginUser(req.Context(), request.Login, request.Password)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !checkUser {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	res.Header().Set("Authorization", "Bearer " + token)
	res.WriteHeader(http.StatusOK)
}

func CreateOrder(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "text/plain" {
		http.Error(res, "Only text/plain content in body are allowed!", http.StatusBadRequest)
		return
	}

	u := security.UserIDType("userID")
	userID := fmt.Sprintf("%v", req.Context().Value(u))
	if userID == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Can't read order number", http.StatusInternalServerError)
		return
	}

	orderNum, err := strconv.ParseInt(string(bodyBytes), 10, 64)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	err = app.ServiceUser.Order(req.Context(), orderNum, userID)
	if errors.Is(err, model.ErrWrongOrderNumber) {
		res.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if errors.Is(err, model.ErrOrderAlreadyExistSameUser) {
		res.WriteHeader(http.StatusOK)
		return
	}

	if errors.Is(err, model.ErrOrderAlreadyExistOtherUser) {
		res.WriteHeader(http.StatusConflict)
		return
	}

	res.WriteHeader(http.StatusAccepted)
}

func CreateWithdrawnOrder(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Only POST requests are allowed!", http.StatusBadRequest)
		return
	}

	u := security.UserIDType("userID")
	userID := fmt.Sprintf("%v", req.Context().Value(u))
	if userID == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	var request model.WithdrawRequest
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&request); err != nil {
		app.Log.Debug("cannot parse request JSON body", zap.Error(err))
		http.Error(res, "cannot parse request JSON body", http.StatusBadRequest)
		return
	}

	app.Log.Debug("Request Create Withdrawn Order", zap.Int64("OrderNum", request.OrderNum), zap.Float32("Sum", request.Sum))

	err := app.ServiceUser.CreateWithdrawnOrder(req.Context(), userID, request.OrderNum, request.Sum)

	if errors.Is(err, model.ErrNoBalance) {
		res.WriteHeader(http.StatusPaymentRequired)
		return
	}

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

func GetBalanceOrders(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	u := security.UserIDType("userID")
	userID := fmt.Sprintf("%v", req.Context().Value(u))
	if userID == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	response, err := app.ServiceUser.GetBalanceOrders(req.Context(), userID)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(response) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	res.WriteHeader(http.StatusOK)
	
	encoder := json.NewEncoder(res)
	if err := encoder.Encode(response); err != nil {
		app.Log.Debug("error encoding response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetWithdrawalsOrders(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	u := security.UserIDType("userID")
	userID := fmt.Sprintf("%v", req.Context().Value(u))
	if userID == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	response, err := app.ServiceUser.GetWithdrawalOrders(req.Context(), userID)
	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(response) == 0 {
		res.WriteHeader(http.StatusNoContent)
		return
	}

	res.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(res)
	if err := encoder.Encode(response); err != nil {
		app.Log.Debug("error encoding response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func GetBalance(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Only GET requests are allowed!", http.StatusBadRequest)
		return
	}

	u := security.UserIDType("userID")
	userID := fmt.Sprintf("%v", req.Context().Value(u))
	if userID == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	response, err := app.ServiceUser.GetBalance(req.Context(), userID)
	if err != nil {
		app.Log.Debug("error get balance", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(res)
	if err := encoder.Encode(response); err != nil {
		app.Log.Debug("error encoding response", zap.Error(err))
		res.WriteHeader(http.StatusInternalServerError)
		return
	}
}
