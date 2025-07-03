package app

import (
	"time"

	"go.uber.org/zap"

	"github.com/kirillmashkov/gophermart/internal/config"
	"github.com/kirillmashkov/gophermart/internal/model"
	"github.com/kirillmashkov/gophermart/internal/service"
	"github.com/kirillmashkov/gophermart/internal/storage"
	"github.com/kirillmashkov/gophermart/internal/util"
)

var ServerConf config.ServerConfig

var Log *zap.Logger = zap.NewNop()
var Database *storage.Database
var ServiceUser *service.ServiceUser
var RepositoryUser *storage.RepositoryUser
var SecurityUtil *util.SecurityUtil

const SecretKey = "supersecretkey"
const tokenExp = time.Hour * 3

func Initialize() error {
	config.InitServerConf(&ServerConf, Log)

	Database = storage.NewDatabase(&ServerConf, Log)
	err := Database.Open()
	if err != nil {
		Log.Error("error open database", zap.Error(err))
		return err
	}

	if err = Database.Migrate(); err != nil {
		return err
	}

	model.OrderNumChanToCreate = make(chan model.OrderToCreate)
	model.RequestToAccrualChan = make(chan model.OrderToRequestAccrual)

	RepositoryUser = storage.NewRepositoryUser(Database, Log)

	SecurityUtil = util.NewSecurityUtil(tokenExp, SecretKey)

	ServiceUser = service.NewServiceUser(Log, RepositoryUser, SecurityUtil, &ServerConf)
	go ServiceUser.CreateBalanceOrder()
	go ServiceUser.RequestOrderAccrual()

	return nil
}

func Close() {
	Database.Close()
}
