package log

import (
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var logger = logrus.New()

type levelFlag struct{}

func (f levelFlag) String() string {
	return logger.Level.String()
}

func (f levelFlag) Set(level string) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logger.Level = l
	return nil
}

func init() {
	formatter := new(prefixed.TextFormatter)
	formatter.DisableTimestamp = true

	logger.Formatter = formatter
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}

func Debugln(args ...interface{}) {
	logger.Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnln(args ...interface{}) {
	logger.Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	logger.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorln(args ...interface{}) {
	logger.Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	logger.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalln(args ...interface{}) {
	logger.Fatalln(args...)
}

func Fatalf(format string, args ...interface{}) {
	logger.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	logger.Panicln(args...)
}

func Panicln(args ...interface{}) {
	logger.Panicln(args...)
}

func Panicf(format string, args ...interface{}) {
	logger.Panicf(format, args...)
}
