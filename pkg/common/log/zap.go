package log

import (
	"context"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/config"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/constant"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/common/tracelog"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	pkgLogger Logger = &ZapLogger{}
	sp               = string(filepath.Separator)
)

// InitFromConfig initializes a Zap-based logger
func InitFromConfig(name string) error {
	l, err := NewZapLogger()
	if err != nil {
		return err
	}
	pkgLogger = l.WithCallDepth(2).WithName(name)
	return nil
}

func ZDebug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	pkgLogger.Debug(ctx, msg, keysAndValues...)
}

func ZInfo(ctx context.Context, msg string, keysAndValues ...interface{}) {
	pkgLogger.Info(ctx, msg, keysAndValues...)
}

func ZWarn(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	pkgLogger.Warn(ctx, msg, err, keysAndValues...)
}

func ZError(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	pkgLogger.Error(ctx, msg, err, keysAndValues...)
}

type ZapLogger struct {
	zap *zap.SugaredLogger
}

func NewZapLogger() (*ZapLogger, error) {
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Encoding:          "json",
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		DisableStacktrace: true,
		InitialFields:     map[string]interface{}{"PID": os.Getegid()},
	}
	if config.Config.Log.Stderr {
		zapConfig.OutputPaths = append(zapConfig.OutputPaths, "stderr")
	}
	zl := &ZapLogger{}
	opts, err := zl.cores()
	if err != nil {
		return nil, err
	}
	l, err := zapConfig.Build(opts)
	if err != nil {
		return nil, err
	}
	zl.zap = l.Sugar()
	zl.zap.WithOptions(zap.AddStacktrace(zap.DPanicLevel))
	return zl, nil
}

func (l *ZapLogger) cores() (zap.Option, error) {
	c := zap.NewProductionEncoderConfig()
	c.EncodeTime = zapcore.ISO8601TimeEncoder
	c.EncodeDuration = zapcore.SecondsDurationEncoder
	c.MessageKey = "msg"
	c.LevelKey = "level"
	c.TimeKey = "time"
	c.CallerKey = "caller"
	//c.EncodeLevel = zapcore.LowercaseColorLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(c)
	fileEncoder.AddInt("PID", os.Getpid())
	writer, err := l.getWriter()
	if err != nil {
		return nil, err
	}
	var cores []zapcore.Core
	if config.Config.Log.StorageLocation != "" {
		cores = []zapcore.Core{
			zapcore.NewCore(fileEncoder, writer, zapcore.DebugLevel),
		}
	}
	if config.Config.Log.Stderr {
		cores = append(cores, zapcore.NewCore(fileEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel))
	}
	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	}), nil
}

func NewErrStackCore(c zapcore.Core) zapcore.Core {
	return &errStackCore{c}
}

type errStackCore struct {
	zapcore.Core
}

func (l *ZapLogger) getWriter() (zapcore.WriteSyncer, error) {
	logf, err := rotatelogs.New(config.Config.Log.StorageLocation+sp+"OpenIM.log.all"+".%Y-%m-%d",
		rotatelogs.WithRotationCount(config.Config.Log.RemainRotationCount),
		rotatelogs.WithRotationTime(time.Duration(config.Config.Log.RotationTime)*time.Hour),
	)
	if err != nil {
		return nil, err
	}
	return zapcore.AddSync(logf), nil
}

func (l *ZapLogger) ToZap() *zap.SugaredLogger {
	return l.zap
}

func (l *ZapLogger) Debug(ctx context.Context, msg string, keysAndValues ...interface{}) {
	keysAndValues = l.kvAppend(ctx, keysAndValues)
	l.zap.Debugw(msg, keysAndValues...)
}

func (l *ZapLogger) Info(ctx context.Context, msg string, keysAndValues ...interface{}) {
	keysAndValues = l.kvAppend(ctx, keysAndValues)
	l.zap.Infow(msg, keysAndValues...)
}

func (l *ZapLogger) Warn(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	if err != nil {
		keysAndValues = append(keysAndValues, "error", err)
	}
	keysAndValues = l.kvAppend(ctx, keysAndValues)
	l.zap.Warnw(msg, keysAndValues...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, err error, keysAndValues ...interface{}) {
	if err != nil {
		keysAndValues = append(keysAndValues, "error", err)
	}
	keysAndValues = append([]interface{}{constant.OperationID, tracelog.GetOperationID(ctx)}, keysAndValues...)
	l.zap.Errorw(msg, keysAndValues...)
}

func (l *ZapLogger) kvAppend(ctx context.Context, keysAndValues []interface{}) []interface{} {
	operationID := tracelog.GetOperationID(ctx)
	opUserID := tracelog.GetOpUserID(ctx)
	connID := tracelog.GetConnID(ctx)
	if opUserID != "" {
		keysAndValues = append([]interface{}{constant.OpUserID, tracelog.GetOpUserID(ctx)}, keysAndValues...)
	}
	if operationID != "" {
		keysAndValues = append([]interface{}{constant.OperationID, tracelog.GetOperationID(ctx)}, keysAndValues...)
	}
	if connID != "" {
		keysAndValues = append([]interface{}{constant.ConnID, tracelog.GetConnID(ctx)}, keysAndValues...)
	}
	return keysAndValues
}

func (l *ZapLogger) WithValues(keysAndValues ...interface{}) Logger {
	dup := *l
	dup.zap = l.zap.With(keysAndValues...)
	return &dup
}

func (l *ZapLogger) WithName(name string) Logger {
	dup := *l
	dup.zap = l.zap.Named(name)
	return &dup
}

func (l *ZapLogger) WithCallDepth(depth int) Logger {
	dup := *l
	dup.zap = l.zap.WithOptions(zap.AddCallerSkip(depth))
	return &dup
}
