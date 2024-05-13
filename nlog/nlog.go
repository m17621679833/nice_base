package nlog

import (
	"fmt"
	"log"
	"path"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	LEVEL_FLAGS = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
)

var (
	defaultLogger *Logger
	takeUp        = false
)

const (
	TRACE = iota
	DEBUG
	INFO
	WARNING
	ERROR
	FATAL
)

const TUNNEL_DEFAULT_SIZE = 1024

type Record struct {
	time  string
	code  string
	info  string
	level int
}

func (r *Record) String() string {
	return fmt.Sprintf("[%s][%s][%s]%s\n", LEVEL_FLAGS[r.level], r.time, r.code, r.info)
}

type Writer interface {
	Init() error
	Write(*Record) error
}

type Rotater interface {
	Rotate() error
	SetPathPattern(string) error
}

type Flusher interface {
	Flush() error
}

type Logger struct {
	writers     []Writer
	tunnel      chan *Record
	level       int
	lastTime    int64
	lastTimeStr string
	c           chan bool
	layout      string
	recordPool  *sync.Pool
}

func (logger *Logger) RegisterWriter(writer Writer) {
	if err := writer.Init(); err != nil {
		panic(err)
	}
	logger.writers = append(logger.writers, writer)
}

func (logger *Logger) SetLogLevel(level int) {
	logger.level = level
}

func (logger *Logger) SetLayout(layout string) {
	logger.layout = layout
}

func (logger *Logger) Trace(fmt string, args ...interface{}) {
	logger.dispatchRecordToTunnel(TRACE, fmt, args...)
}

func (logger *Logger) Debug(fmt string, args ...interface{}) {
	logger.dispatchRecordToTunnel(DEBUG, fmt, args...)
}

func (logger *Logger) Info(fmt string, args ...interface{}) {
	logger.dispatchRecordToTunnel(INFO, fmt, args...)
}
func (logger *Logger) Warn(fmt string, args ...interface{}) {
	logger.dispatchRecordToTunnel(WARNING, fmt, args...)
}
func (logger *Logger) Error(fmt string, args ...interface{}) {
	logger.dispatchRecordToTunnel(ERROR, fmt, args...)
}
func (logger *Logger) Fatal(fmt string, args ...interface{}) {
	logger.dispatchRecordToTunnel(FATAL, fmt, args...)
}

func (logger *Logger) Close() {
	close(logger.tunnel)
	<-logger.c
	for _, writer := range logger.writers {
		if flusher, ok := writer.(Flusher); ok {
			if err := flusher.Flush(); err != nil {
				log.Println(err)
			}
		}
	}
}

/*将logger中sync.pool中的日志记录分发到日志tunnel中*/
func (logger *Logger) dispatchRecordToTunnel(level int, format string, args ...interface{}) {
	var inf, code string
	if level < logger.level {
		return
	}
	if format != "" {
		inf = fmt.Sprintf(format, args...)
	} else {
		inf = fmt.Sprint(args...)
	}
	_, file, line, ok := runtime.Caller(2)
	if ok {
		code = path.Base(file) + ":" + strconv.Itoa(line)
	}
	now := time.Now()
	if now.Unix() != logger.lastTime {
		logger.lastTime = now.Unix()
		logger.lastTimeStr = now.Format(logger.layout)
	}
	record := logger.recordPool.Get().(*Record)
	record.info = inf
	record.code = code
	record.time = logger.lastTimeStr
	record.level = level
	logger.tunnel <- record
}

func bootstrapLogWriter(logger *Logger) {
	if logger == nil {
		panic("logger is nil")
	}

	var (
		r  *Record
		ok bool
	)

	if r, ok = <-logger.tunnel; !ok {
		logger.c <- true
		return
	}

	for _, writer := range logger.writers {
		if err := writer.Write(r); err != nil {
			log.Println(err)
		}
	}

	flushTimer := time.NewTimer(time.Millisecond * 500)
	rotateTimer := time.NewTimer(time.Millisecond * 10)
	for {
		select {
		case r, ok = <-logger.tunnel:
			if !ok {
				logger.c <- true
				return
			}
			for _, writer := range logger.writers {
				if err := writer.Write(r); err != nil {
					log.Println(err)
				}
			}
			logger.recordPool.Put(r)
		case <-flushTimer.C:
			for _, writer := range logger.writers {
				if flusher, ok := writer.(Flusher); ok {
					if err := flusher.Flush(); err != nil {
						log.Println(err)
					}
				}
			}
			flushTimer.Reset(time.Millisecond * 1000)
		case <-rotateTimer.C:
			for _, writer := range logger.writers {
				if rotater, ok := writer.(Rotater); ok {
					if err := rotater.Rotate(); err != nil {
						log.Println(err)
					}
				}
			}
			rotateTimer.Reset(time.Second * 10)
		}
	}
}

func NewLogger() *Logger {
	if defaultLogger != nil && takeUp == false {
		takeUp = true
		return defaultLogger
	}
	logger := new(Logger)
	logger.writers = []Writer{}
	logger.tunnel = make(chan *Record, TUNNEL_DEFAULT_SIZE)
	logger.c = make(chan bool, 2)
	logger.level = DEBUG
	logger.layout = "2006/01/02 15:04:05"

	logger.recordPool = &sync.Pool{
		New: func() interface{} { return &Record{} },
	}
	go bootstrapLogWriter(logger)
	return logger
}

func InitDefaultLogger() {
	if takeUp == false {
		defaultLogger = NewLogger()
	}
}

func SetLayout(layout string) {
	InitDefaultLogger()
	defaultLogger.layout = layout
}

func SetLogLevel(level int) {
	InitDefaultLogger()
	defaultLogger.level = level
}

func Trace(fmt string, args ...interface{}) {
	InitDefaultLogger()
	defaultLogger.dispatchRecordToTunnel(TRACE, fmt, args...)
}

func Debug(fmt string, args ...interface{}) {
	InitDefaultLogger()
	defaultLogger.dispatchRecordToTunnel(DEBUG, fmt, args...)
}

func Info(fmt string, args ...interface{}) {
	InitDefaultLogger()
	defaultLogger.dispatchRecordToTunnel(INFO, fmt, args...)
}

func Warn(fmt string, args ...interface{}) {
	InitDefaultLogger()
	defaultLogger.dispatchRecordToTunnel(WARNING, fmt, args...)
}

func Error(fmt string, args ...interface{}) {
	InitDefaultLogger()
	defaultLogger.dispatchRecordToTunnel(ERROR, fmt, args...)
}
func Fatal(fmt string, args ...interface{}) {
	InitDefaultLogger()
	defaultLogger.dispatchRecordToTunnel(FATAL, fmt, args...)
}

func Register(writer Writer) {
	InitDefaultLogger()
	defaultLogger.RegisterWriter(writer)
}

func Close() {
	defaultLogger.Close()
	defaultLogger = nil
	takeUp = false
}
