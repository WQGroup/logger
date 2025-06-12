package logger

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestAll(t *testing.T) {

	SetLoggerOnlyMsg(true)

	SetLoggerLevel(logrus.DebugLevel)

	SetLoggerRootDir(".")

	SetLoggerName("TestAppLogger")

	formatString := "%s"
	srcFormatString := "haha"
	Debugf(formatString, srcFormatString)
	Infof(formatString, srcFormatString)

	println(LogLinkFileFPath())
	println(CurrentFileName())
}
