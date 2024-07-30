package loggers

import (
	"github.com/Roh-bot/rabbitmq-pub-sub/internal/config"
	"github.com/Roh-bot/rabbitmq-pub-sub/pkg/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
)

var Zap = &logger{}

type logger struct {
	*zap.SugaredLogger
}

func ZapNew(cores ...zapcore.Core) error {
	stdoutLogger := StdoutLogger()
	stdoutCore := NewZapCore(zapcore.Lock(stdoutLogger), config.GetConfig().Logger.Level)

	// Combine the cores if provided
	var tee zapcore.Core
	if global.IsDevelopment {
		tee = zapcore.NewTee(
			append([]zapcore.Core{stdoutCore}, cores...)...,
		)
	} else {
		if len(cores) > 0 {
			tee = zapcore.NewTee(
				cores...,
			)
		}
	}

	z := zap.New(
		tee,
		zap.AddCaller(),
		zap.AddCallerSkip(1), // Skip one level to get the correct caller
	).
		Sugar()
	zap.Fields()
	Zap.SugaredLogger = z
	return nil
}

func LogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dPanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

func CustomEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		MessageKey:     "tag",
		LevelKey:       "level",
		TimeKey:        "timestamp",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func NewZapCore(syncer zapcore.WriteSyncer, level string) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(CustomEncoderConfig()),
		syncer,
		LogLevel(level))
}

func NewZapCoreLevelEnabler(syncer zapcore.WriteSyncer, level string) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(CustomEncoderConfig()),
		syncer,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl == LogLevel(level)
		}))
}

func NewZapCoreLevelsEnabler(syncer zapcore.WriteSyncer, level string) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(CustomEncoderConfig()),
		syncer,
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl <= LogLevel(level)
		}))
}

func (l *logger) Sync() {
	// Sync the logger
	if err := l.SugaredLogger.Sync(); err != nil {
		log.Print(err)
		return
	}
	return
}

func (l *logger) Errorln(args ...any) {
	if args == nil || len(args) == 0 {
		l.SugaredLogger.Errorln(args)
		return
	}
	l.SugaredLogger.Errorln(args)
}

func (l *logger) Warnln(args ...any) {
	if args == nil || len(args) == 0 {
		l.SugaredLogger.Warnln(args)
		return
	}
	l.SugaredLogger.Warnln(args)
}

func (l *logger) Infoln(args ...any) {
	if args == nil || len(args) == 0 {
		l.SugaredLogger.Infoln(args)
		return
	}
	l.SugaredLogger.Infoln(args)
}

func (l *logger) Debugln(args ...any) {
	if args == nil || len(args) == 0 {
		l.SugaredLogger.Debugln(args)
		return
	}
	l.SugaredLogger.Debugln(args)
}

func (l *logger) Errorf(template string, args ...any) {
	if args == nil || len(args) == 0 {
		l.SugaredLogger.Errorf(template)
		return
	}
	l.SugaredLogger.Errorf(template, args)
}

func (l *logger) Warnf(template string, args ...any) {
	if args == nil || len(args) == 0 {
		l.SugaredLogger.Warnf(template)
		return
	}
	l.SugaredLogger.Warnf(template, args)
}

func (l *logger) Infof(template string, args ...any) {
	if args == nil || len(args) == 0 {
		l.SugaredLogger.Infof(template)
		return
	}
	l.SugaredLogger.Infof(template, args)
}

func (l *logger) Debugf(template string, args ...any) {
	if args == nil || len(args) == 0 {
		l.SugaredLogger.Debugf(template)
		return
	}
	l.SugaredLogger.Debugf(template, args)
}
