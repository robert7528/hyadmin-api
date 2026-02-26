package logger

import (
	"github.com/hysp/hyadmin-api/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

func New(cfg *config.Config) (*zap.Logger, error) {
	level := zap.InfoLevel
	if cfg.Log.Level == "debug" {
		level = zap.DebugLevel
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "time"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Log.Filename,
		MaxSize:    100, // MB
		MaxBackups: 3,
		MaxAge:     28, // days
		Compress:   true,
	})
	consoleWriter := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), fileWriter, level),
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderCfg), consoleWriter, level),
	)
	return zap.New(core, zap.AddCaller()), nil
}
