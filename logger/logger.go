package logger

import (
	"context"
	"fsn"
	"io"
	"path"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
)

func NewLogger(ctx context.Context, out io.Writer) context.Context {
	var logger = &logrus.Logger{
		Out:       out,
		Formatter: new(logrus.TextFormatter),
		Hooks:     make(logrus.LevelHooks),
		Level:     logrus.DebugLevel,
	}

	return WithLogger(ctx, logrus.NewEntry(logger))
}

func addSpaces(args ...interface{}) []interface{} {
	res := make([]interface{}, 0, len(args))

	for i, arg := range args {
		if (i+1)%2 != 0 && len(args) > 1 {
			str, ok := arg.(string)
			if ok {
				str += " "
				arg = str
			}
		}

		res = append(res, arg)
	}

	return res
}

func Info(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Info(addSpaces(args)...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Infof(format, args...)
}

func Debug(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Debug(addSpaces(args)...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Debugf(format, args...)
}

func Warn(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Warn(addSpaces(args)...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Warnf(format, args...)
}

func Error(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Error(addSpaces(args)...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Errorf(format, args...)
}

func Trace(ctx context.Context, args ...interface{}) {
	logger := ctxlogrus.Extract(ctx)

	logger.Trace(addSpaces(args...))
}

func WithLogger(ctx context.Context, logger *logrus.Entry) context.Context {
	return ctxlogrus.ToContext(ctx, logger)
}

func GetFilepath() string {
	return path.Join(fsn.Root, "log", time.Now().Format("02.01.2006_15:04"))
}
