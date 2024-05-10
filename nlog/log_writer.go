package nlog

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"time"
)

var pathVariableMap map[byte]func(time *time.Time) int

type LogWriter struct {
	minLogLevel   int
	maxLogLevel   int
	fileName      string
	pathFmt       string
	file          *os.File
	fileBufWriter *bufio.Writer
	actions       []func(*time.Time) int
	variables     []interface{}
}

func NewLogWriter() *LogWriter {
	return &LogWriter{}
}

func (logWriter *LogWriter) Init() error {
	return logWriter.CreateLogFile()
}

func (logWriter *LogWriter) SetFileName(fileName string) {
	logWriter.fileName = fileName
}

func (logWriter *LogWriter) SetMinLogLevel(min int) {
	logWriter.minLogLevel = min
}

func (logWriter *LogWriter) SetMaxLogLevel(max int) {
	logWriter.maxLogLevel = max
}

func (logWriter *LogWriter) SetPathPattern(pattern string) error {
	n := 0
	for _, c := range pattern {
		if c == '%' {
			n++
		}
	}

	if n == 0 {
		logWriter.pathFmt = pattern
		return nil
	}

	logWriter.actions = make([]func(*time.Time) int, 0, n)
	logWriter.variables = make([]interface{}, n, n)
	tmp := []byte(pattern)

	variable := 0
	for _, c := range tmp {
		if variable == 1 {
			act, ok := pathVariableMap[c]
			if !ok {
				return errors.New("Invalid rotate pattern(" + pattern + ")")
			}
			logWriter.actions = append(logWriter.actions, act)
		}
		if c == '%' {
			variable = 1
		}
	}
	for i, action := range logWriter.actions {
		now := time.Now()
		logWriter.variables[i] = action(&now)
	}
	logWriter.pathFmt = convertPatternToFmt(tmp)
	return nil
}

func (logWriter *LogWriter) Write(record *Record) error {
	if record.level < logWriter.minLogLevel || record.level > logWriter.maxLogLevel {
		return nil
	}
	if logWriter.fileBufWriter == nil {
		return errors.New("no opened file")
	}
	if _, err := logWriter.fileBufWriter.WriteString(record.String()); err != nil {
		return err
	}
	return nil
}

func (logWriter *LogWriter) Rotate() error {
	now := time.Now()
	v := 0
	rotate := false
	oldVariable := make([]interface{}, len(logWriter.variables))
	copy(oldVariable, logWriter.variables)
	for i, action := range logWriter.actions {
		v = action(&now)
		if v != logWriter.variables[i] {
			logWriter.variables[i] = v
			rotate = true
		}
	}

	if rotate == false {
		return nil
	}

	if logWriter.fileBufWriter != nil {
		if err := logWriter.fileBufWriter.Flush(); err != nil {
			return err
		}
	}

	if logWriter.file != nil {
		filePath := fmt.Sprintf(logWriter.pathFmt, oldVariable...)
		if err := os.Rename(logWriter.fileName, filePath); err != nil {
			return err
		}
		if err := logWriter.file.Close(); err != nil {
			return err
		}
	}

	return logWriter.CreateLogFile()
}

func convertPatternToFmt(pattern []byte) string {
	pattern = bytes.Replace(pattern, []byte("%Y"), []byte("%d"), -1)
	pattern = bytes.Replace(pattern, []byte("%M"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%D"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%H"), []byte("%02d"), -1)
	pattern = bytes.Replace(pattern, []byte("%m"), []byte("%02d"), -1)
	return string(pattern)
}

func (logWriter *LogWriter) CreateLogFile() error {
	if err := os.MkdirAll(path.Dir(logWriter.fileName), 0755); err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	if file, err := os.OpenFile(logWriter.fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644); err != nil {
		return err
	} else {
		logWriter.file = file
	}

	if logWriter.fileBufWriter = bufio.NewWriterSize(logWriter.file, 8192); logWriter.fileBufWriter == nil {
		return errors.New("new fileBufWriter failed")
	}
	return nil
}

func (logWriter *LogWriter) Flush() error {
	if logWriter.fileBufWriter != nil {
		return logWriter.fileBufWriter.Flush()
	}
	return nil
}

func getYear(now *time.Time) int {
	return now.Year()
}

func getMonth(now *time.Time) int {
	return int(now.Month())
}

func getDay(now *time.Time) int {
	return now.Day()
}

func getHour(now *time.Time) int {
	return now.Hour()
}

func getMin(now *time.Time) int {
	return now.Minute()
}

func init() {
	pathVariableMap = make(map[byte]func(*time.Time) int, 5)
	pathVariableMap['Y'] = getYear
	pathVariableMap['M'] = getMonth
	pathVariableMap['D'] = getDay
	pathVariableMap['H'] = getHour
	pathVariableMap['m'] = getMin
}
