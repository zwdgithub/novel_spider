package log

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var dbLogger *zap.SugaredLogger

func init() {

	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:  "log/db.log",
		MaxSize:   100, //MB
		LocalTime: true,
		Compress:  true,
	})
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(config),
		w,
		logLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	dbLogger = logger.Sugar()
}

type DBLogger struct {
}

func (l DBLogger) Print(values ...interface{}) {
	dbLogger.Info(gorm.LogFormatter(values...)...)
}
