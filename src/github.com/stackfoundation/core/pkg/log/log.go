package log

import "github.com/Sirupsen/logrus"

func SetDebug(debug bool) {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func Debugf(format string, args ...interface{}) {
	logrus.Debugf(format, args...)
}
