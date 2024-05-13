package test

import (
	"github.com/m17621679833/nice_base/nlog"
	"testing"
	"time"
)

func TestNLog(t *testing.T) {
	logger := nlog.NewLogger()
	config := &nlog.LogConfig{
		LogLevel: "trace",
		FileWriter: nlog.FileWriterConf{
			On:              true,
			LogPath:         "./log_test.log",
			RotateLogPath:   "./log_test.log",
			WfLogPath:       "./log_test.wf.log",
			RotateWfLogPath: "./log_test.wf.log",
		},
		ConsoleWriter: nlog.ConsoleWriterConf{
			On:    true,
			Color: true,
		},
	}
	nlog.SetupLogInstanceWithConf(config, logger)

	time.Sleep(12 * time.Second)

	logger.Error("2error message")
	logger.Close()
	time.Sleep(60 * time.Second)
}
