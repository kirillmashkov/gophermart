package model

import "time"

type LoginPassRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type OrderToCreate struct {
	OrderNum int64
	UserID   string
}

type OrderToRequestAccrual struct {
	OrderID  string
	OrderNum int64
	UserID   string
}

type WithdrawRequest struct {
	OrderNum int64 `json:"order"`
	Sum float32 `json:"sum"`
}

type AccrualResponse struct {
	Order   string `json:"order"`
	Status  string `json:"status"`
	Accrual float32    `json:"accrual"`
}

type OrdersDB struct {
	Number     string
	Status     string
	Sum        float32
	UploadedAt time.Time
}

type OrdersBalanceReposponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float32       `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type OrderWithdrawlResponse struct {
	Order       string    `json:"order"`
	Sum         float32       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type BalanceResponse struct {
	Balance   float32 `json:"current"`
	Withdrawn float32 `json:"withdrawn"`
}

var OrderNumChanToCreate chan OrderToCreate
var RequestToAccrualChan chan OrderToRequestAccrual
