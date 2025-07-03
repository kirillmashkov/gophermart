package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

type ServerConfig struct {
	Host        		string "env:\"RUN_ADDRESS\""
	AccrualAddress    	string "env:\"ACCRUAL_SYSTEM_ADDRESS\""
	Connection			string "env:\"DATABASE_URI\""
}

var ServerEnv ServerConfig
var ServerArg ServerConfig

func init() {
	flag.StringVar(&ServerArg.Host, "a", "localhost:8081", "server host")
	flag.StringVar(&ServerArg.Connection, "d", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable", "db connection string")
	flag.StringVar(&ServerArg.AccrualAddress, "r", "http://localhost:8080", "accrual system address")
}

func InitServerConf(conf *ServerConfig, logger *zap.Logger) {
	err := env.Parse(&ServerEnv)
	if err != nil {
		logger.Error("Can't read env variables")
	}

	conf.AccrualAddress = getConfigString(ServerEnv.AccrualAddress, ServerArg.AccrualAddress)
	conf.Host = getConfigString(ServerEnv.Host, ServerArg.Host)
	conf.Connection = getConfigString(ServerEnv.Connection, ServerArg.Connection)

	logger.Info("server config",
		zap.String("host", conf.Host),
		zap.String("redirect", conf.AccrualAddress),
		zap.String("db connection", conf.Connection))
}

func getConfigString(env string, arg string) string {
	if env == "" {
		return arg
	} else {
		return env
	}
}