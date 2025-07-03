package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/kirillmashkov/gophermart/internal/app"
	"github.com/kirillmashkov/gophermart/internal/httpserver/middleware/compress"
	middlewarelogger "github.com/kirillmashkov/gophermart/internal/httpserver/middleware/logger"
	"github.com/kirillmashkov/gophermart/internal/httpserver/middleware/security"
	"github.com/kirillmashkov/gophermart/internal/logger"
	"github.com/kirillmashkov/gophermart/internal/model"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func setDefaultLog() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	log.SetPrefix("ERROR: ")
	log.Println("Can't init logger")
}

func initLogger() {
	err := logger.Initialize()
	if err != nil {
		setDefaultLog()
		panic(err)
	}
}

func initApp() {
	err := app.Initialize()
	if err != nil {
		app.Log.Error("Error init app", zap.Error(err))
		panic(err)
	}
}

func initServ() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middlewarelogger.Logger)
	r.Use(compress.Compress)
	r.Use(security.Auth)
	return r
}

func changeWorkingDir() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../../../")
	err := os.Chdir(dir)
  	if err != nil {
		setDefaultLog()
    	panic(err)
 	}
}

func TestRegisterUser(t *testing.T) {
	changeWorkingDir()
	initLogger()
	initApp()

	type want struct {
		code int
	}

	tests := []struct {
		name string
		request model.LoginPassRequest
		want want

	} {
		{
			name: "Register User",
			request: model.LoginPassRequest {
				Login: "admin",
				Password: "admin",
			},
			want: want {
				code: 200,
			},
		},

		{
			name: "Login already used",
			request: model.LoginPassRequest {
				Login: "admin",
				Password: "admin",
			},
			want: want {
				code: 409,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := register(test.request)
			assert.Equal(t, test.want.code, res.StatusCode)

			if res.StatusCode == 200 {
				checkToken(t, *res, test.request.Login)
			}

			if errClose := res.Body.Close(); errClose != nil {
				app.Log.Error("Can't close", zap.Error(errClose))
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	changeWorkingDir()
	initLogger()
	initApp()

	type want struct {
		code int
	}

	tests := []struct {
		name string
		request model.LoginPassRequest
		want want

	} {
		{
			name: "Login User",
			request: model.LoginPassRequest {
				Login: "admin",
				Password: "admin",
			},
			want: want {
				code: 200,
			},
		},
		{
			name: "Fail login user",
			request: model.LoginPassRequest {
				Login: "admin",
				Password: "admin1",
			},
			want: want {
				code: 401,
			},
		},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := login(test.request)

			assert.Equal(t, test.want.code, res.StatusCode)

			if res.StatusCode == 200 {
				checkToken(t, *res, test.request.Login)
			}

			if errClose := res.Body.Close(); errClose != nil {
				app.Log.Error("Can't close", zap.Error(errClose))
			}

		})
	}
}

func checkToken(t *testing.T, res http.Response, login string) {
	tokenHeader := res.Header.Get("Authorization")
	tokenString := (strings.Split(tokenHeader, "Bearer "))[1]
	app.Log.Info("Token", zap.String("token", tokenString))
	token, claims, err := app.SecurityUtil.ParseJWT(tokenString)
	if err != nil {
		app.Log.Error("Can't parse token", zap.Error(err))
		panic(err)
	}

	assert.Equal(t, true, token.Valid)
	
	var id string
	err = app.Database.Dbpool.QueryRow(context.TODO(), "select id from profile where login = $1", login).Scan(&id)
	if err != nil {
		app.Log.Error("Can't read from DB", zap.Error(err))
		panic(err)
	}
	assert.Equal(t, id, claims.UserID)

}

func register(req model.LoginPassRequest) *http.Response {
	requestBytes, err := json.Marshal(req)
	if err != nil {
		app.Log.Error("Error marshal request", zap.Error(err))
		panic(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(requestBytes))
	w := httptest.NewRecorder()
	RegisterUser(w, request)
	return w.Result()
}

func login(req model.LoginPassRequest) *http.Response {
	requestBytes, err := json.Marshal(req)
	if err != nil {
		app.Log.Error("Error marshal auth request", zap.Error(err))
		panic(err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(requestBytes))
	w := httptest.NewRecorder()
	LoginUser(w, request)
	return w.Result()
}

func TestCreateOrder(t *testing.T) {
	changeWorkingDir()
	initLogger()
	initApp()
	r := initServ()
	r.Post("/api/user/orders", CreateOrder)

	type requestType struct {
		auth bool
		requestAuth model.LoginPassRequest
		contentType string
		orderNum string
	}

	type want struct {
		code int
	}

	tests := []struct {
		name string
		request requestType
		want
	} {
		{
			name: "Create Order",
			request: requestType {
				auth: true,
				requestAuth: model.LoginPassRequest {
					Login: "admin1",
					Password: "admin1",
				},
				contentType: "text/plain",
				orderNum: "12345678903",
			},
			want: want {
				code: 202,
			},
		},
		{
			name: "Wrong num order",
			request: requestType {
				auth: true,
				requestAuth: model.LoginPassRequest {
					Login: "admin2",
					Password: "admin2",
				},
				contentType: "text/plain",
				orderNum: "12345678902",
			},
			want: want {
				code: 422,
			},
		},
		{
			name: "Unauthorized",
			request: requestType {
				auth: false,
				requestAuth: model.LoginPassRequest {
					Login: "",
					Password: "",
				},
				contentType: "text/plain",
				orderNum: "12345678903",
			},
			want: want {
				code: 401,
			},
		},
		{
			name: "Order already has been uploaded by other user",
			request: requestType {
				auth: true,
				requestAuth: model.LoginPassRequest {
					Login: "admin3",
					Password: "admin3",
				},
				contentType: "text/plain",
				orderNum: "12345678903",
			},
			want: want {
				code: 409,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var tokenJWT = ""
			if test.request.auth {
				resAuth := register(test.request.requestAuth)
				tokenJWT = resAuth.Header.Get("Authorization")
			}

			app.Log.Info("token", zap.String("token", tokenJWT))
			request := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(test.request.orderNum))
			request.Header.Set("Authorization", tokenJWT)
			request.Header.Set("Content-Type", test.request.contentType)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, test.want.code, res.StatusCode)

			if errClose := res.Body.Close(); errClose != nil {
				app.Log.Error("Can't close", zap.Error(errClose))
			}

		})
	}
}

