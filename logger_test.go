package logger

import (
	"testing"
)

func TestAll(t *testing.T) {

	settings := NewSettings()
	settings.OnlyMsg = true
	settings.LogRootFPath = "./mylogs"
	SetLoggerSettings(settings)

	formatString := "%s"
	srcFormatString := "haha"
	Info(srcFormatString)
	Debugf(formatString, srcFormatString)
	Infof(formatString, srcFormatString)

	println(LogLinkFileFPath())
	println(CurrentFileName())
}
