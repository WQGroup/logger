package logger

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestAll(t *testing.T) {

	SetLoggerLevel(logrus.DebugLevel)

	SetLoggerRootDir(".")

	SetLoggerName("TestAppLogger")

	formatString := "%s"
	srcFormatString := "haha"
	Debugf(formatString, srcFormatString)
	Infof(formatString, srcFormatString)
}
