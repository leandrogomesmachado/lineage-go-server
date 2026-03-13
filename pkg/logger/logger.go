package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.SugaredLogger

func Init(level string, logFile string, console bool) error {
	var cores []zapcore.Core

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	logLevel := zapcore.InfoLevel
	switch level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	}

	if console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel)
		cores = append(cores, consoleCore)
	}

	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(file), logLevel)
		cores = append(cores, fileCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	Log = logger.Sugar()

	return nil
}

func Info(args ...interface{}) {
	if Log != nil {
		Log.Info(args...)
	}
}

func Infof(template string, args ...interface{}) {
	if Log != nil {
		Log.Infof(template, args...)
	}
}

func Debug(args ...interface{}) {
	if Log != nil {
		Log.Debug(args...)
	}
}

func Debugf(template string, args ...interface{}) {
	if Log != nil {
		Log.Debugf(template, args...)
	}
}

func Warn(args ...interface{}) {
	if Log != nil {
		Log.Warn(args...)
	}
}

func Warnf(template string, args ...interface{}) {
	if Log != nil {
		Log.Warnf(template, args...)
	}
}

func Error(args ...interface{}) {
	if Log != nil {
		Log.Error(args...)
	}
}

func Errorf(template string, args ...interface{}) {
	if Log != nil {
		Log.Errorf(template, args...)
	}
}

func Fatal(args ...interface{}) {
	if Log != nil {
		Log.Fatal(args...)
	}
}

func Fatalf(template string, args ...interface{}) {
	if Log != nil {
		Log.Fatalf(template, args...)
	}
}
