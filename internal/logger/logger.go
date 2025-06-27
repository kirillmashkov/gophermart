package logger

import (
	"os"

	"github.com/kirillmashkov/gophermart/internal/app"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

type Settings struct {
	Logger struct {
		Level string `yaml:"level"`
		Development bool `yaml:"development"`
		DisableCaller bool `yaml:"disable-caller"`
		DisableStacktrace bool `yaml:"disable-stacktracce"`
		Encoding string `yaml:"encoding"`
		OutputPaths []string `yaml:"output-paths"`
		ErrorOutputPaths []string `yaml:"error-output-paths"`

	} `yaml:"logger"`
}

const filenameConfig = "config/config.yml" 

func Initialize() error {
	encoderCfg := zap.NewProductionEncoderConfig()
    encoderCfg.TimeKey = "timestamp"
    encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	data, err := os.ReadFile(filenameConfig)
	if err != nil {
		return err
	}

	var settings Settings
	err = yaml.Unmarshal(data, &settings)
	if err != nil {
		return err
	}

    config := zap.Config{
        Level:             	zap.NewAtomicLevelAt(getLevelLogger(settings.Logger.Level)),
        Development:       	settings.Logger.Development,
        DisableCaller:     	settings.Logger.DisableCaller,
        DisableStacktrace: 	settings.Logger.DisableStacktrace,
        Sampling:          	nil,
        Encoding:          	settings.Logger.Encoding,
        EncoderConfig:     	encoderCfg,
        OutputPaths: 		settings.Logger.OutputPaths,
        ErrorOutputPaths: 	settings.Logger.ErrorOutputPaths,
        InitialFields: map[string]interface{}{
            "pid": os.Getpid(),
        },
    }
	app.Log = zap.Must(config.Build()) 
	return nil
}

func getLevelLogger(level string) zapcore.Level {
	switch level {
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	case "warn":
		return zap.WarnLevel
	default:
		return zap.ErrorLevel
	}
}
