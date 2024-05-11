package nlog

import "errors"

type FileWriterConf struct {
	On              bool   `toml:"On"`
	LogPath         string `toml:"LogPath"`
	RotateLogPath   string `toml:"RotateLogPath"`
	WfLogPath       string `toml:"WfLogPath"`
	RotateWfLogPath string `toml:"RotateWfLogPath"`
}

type ConsoleWriterConf struct {
	On    bool `toml:"On"`
	Color bool `toml:"Color"`
}

type LogConfig struct {
	LogLevel      string            `toml:"LogLevel"`
	FileWriter    FileWriterConf    `toml:"FileWriter"`
	ConsoleWriter ConsoleWriterConf `toml:"ConsoleWriter"`
}

func SetupLogInstanceWithConf(lc LogConfig, logger *Logger) (err error) {
	if lc.FileWriter.On {
		if len(lc.FileWriter.LogPath) > 0 {
			w := NewLogWriter()
			w.SetFileName(lc.FileWriter.LogPath)
			w.SetPathPattern(lc.FileWriter.RotateLogPath)
			w.SetMinLogLevel(TRACE)
			if len(lc.FileWriter.WfLogPath) > 0 {
				w.SetMaxLogLevel(INFO)
			} else {
				w.SetMaxLogLevel(ERROR)
			}
			logger.RegisterWriter(w)
		}

		if len(lc.FileWriter.WfLogPath) > 0 {
			wfw := NewLogWriter()
			wfw.SetFileName(lc.FileWriter.WfLogPath)
			wfw.SetPathPattern(lc.FileWriter.RotateWfLogPath)
			wfw.SetMinLogLevel(WARNING)
			wfw.SetMaxLogLevel(ERROR)
			logger.RegisterWriter(wfw)
		}
	}

	if lc.ConsoleWriter.On {
		w := NewConsoleWriter()
		w.SetColor(lc.ConsoleWriter.Color)
		logger.RegisterWriter(w)
	}
	switch lc.LogLevel {
	case "trace":
		logger.SetLogLevel(TRACE)

	case "debug":
		logger.SetLogLevel(DEBUG)

	case "info":
		logger.SetLogLevel(INFO)

	case "warning":
		logger.SetLogLevel(WARNING)

	case "error":
		logger.SetLogLevel(ERROR)

	case "fatal":
		logger.SetLogLevel(FATAL)

	default:
		err = errors.New("Invalid log level")
	}
	return
}

func SetupDefaultLogWithConf(lc LogConfig) (err error) {
	InitDefaultLogger()
	return SetupLogInstanceWithConf(lc, defaultLogger)
}
