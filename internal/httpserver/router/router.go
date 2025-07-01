package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kirillmashkov/gophermart/internal/httpserver/handler"
	// "github.com/kirillmashkov/gophermart/internal/httpserver/middleware/compress"
	"github.com/kirillmashkov/gophermart/internal/httpserver/middleware/logger"
	"github.com/kirillmashkov/gophermart/internal/httpserver/middleware/security"
	"github.com/go-chi/chi/middleware"
)

func Serv() http.Handler {
	r := chi.NewRouter()
	r.Use(logger.Logger)
	// r.Use(compress.Compress)
	r.Use(middleware.Compress(5, "gzip"))
	r.Use(security.Auth)

	r.Post("/api/user/register", handler.RegisterUser)
	r.Post("/api/user/login", handler.LoginUser)
	r.Post("/api/user/orders", handler.CreateOrder)
	r.Get("/api/user/orders", handler.GetBalanceOrders)
	r.Get("/api/user/withdrawals", handler.GetWithdrawalsOrders)
	r.Get("/api/user/balance", handler.GetBalance)
	r.Post("/api/user/balance/withdraw", handler.CreateWithdrawnOrder)
	// r.Post("/api/shorten/batch", handler.PostGenerateShortURLBatch)
	// r.Delete("/api/user/urls", handler.DeleteURLBatch)
	// r.Get("/ping", handler.Ping)

	return r
}
