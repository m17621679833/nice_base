package lib

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

func InitDBPool(path string) error {
	dbConfMap := &MysqlConfMap{}
	err := ParseConfig(path, dbConfMap)
	if err != nil {
		return err
	}
	if len(dbConfMap.List) == 0 {
		fmt.Printf("[INFO]%s%s\n", time.Now().Format(TimeFormat), "empty mysql config")
		return errors.New("初始化mysql失败~！")
	}
	DBMapPool = map[string]*sql.DB{}
	GORMMapPool = map[string]*gorm.DB{}
	for confName, conf := range dbConfMap.List {
		dbPool, err := sql.Open("mysql", conf.DataSourceName)
		if err != nil {
			return err
		}
		dbPool.SetMaxOpenConns(conf.MaxOpenConn)
		dbPool.SetMaxIdleConns(conf.MaxIdleConn)
		dbPool.SetConnMaxLifetime(time.Duration(conf.MaxConnLifeTime) * time.Second)
		err = dbPool.Ping()
		dbGorm, err := gorm.Open(mysql.New(mysql.Config{Conn: dbPool}), &gorm.Config{
			Logger: &DefaultMysqlGormLogger,
		})
		if err != nil {
			return err
		}
		DBMapPool[confName] = dbPool
		GORMMapPool[confName] = dbGorm
	}

	if pool, err := GetDBPool("default"); err == nil {
		DBDefaultPool = pool
	}

	if pool, err := GetGormPool("default"); err == nil {
		GORMDefaultPool = pool
	}
	return nil
}

func GetDBPool(name string) (*sql.DB, error) {
	if pool, ok := DBMapPool[name]; ok {
		return pool, nil
	}
	return nil, errors.New("get pool error")
}

func GetGormPool(name string) (*gorm.DB, error) {
	if pool, ok := GORMMapPool[name]; ok {
		return pool, nil
	}
	return nil, errors.New("get gorm pool error")
}

// mysql 日志打印类型
var DefaultMysqlGormLogger = MysqlGormLogger{
	LogLevel:      logger.Info,
	SlowThreshold: 200 * time.Millisecond,
}

type MysqlGormLogger struct {
	LogLevel      logger.LogLevel
	SlowThreshold time.Duration
}

func (m MysqlGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	m.LogLevel = level
	return m
}

func (m MysqlGormLogger) Info(ctx context.Context, s string, i ...interface{}) {
	traceContext := GetTraceContext(ctx)
	params := make(map[string]interface{})
	params["message"] = s
	params["values"] = fmt.Sprint(i...)
	Log.TagInfo(traceContext, "_com_mysql_Info", params)
}

func (m MysqlGormLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	traceContext := GetTraceContext(ctx)
	params := make(map[string]interface{})
	params["message"] = s
	params["values"] = i
	Log.TagInfo(traceContext, "_com_mysql_Warn", params)
}

func (m MysqlGormLogger) Error(ctx context.Context, message string, values ...interface{}) {
	trace := GetTraceContext(ctx)
	params := make(map[string]interface{})
	params["message"] = message
	params["values"] = fmt.Sprint(values...)
	Log.TagInfo(trace, "_com_mysql_Error", params)
}

func (m MysqlGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	traceContext := GetTraceContext(ctx)
	if m.LogLevel <= logger.Silent {
		return
	}
	sqlStr, rows := fc()
	currentTime := begin.Format(TimeFormat)
	since := time.Since(begin)
	msg := map[string]interface{}{
		"FileWithLineNum": utils.FileWithLineNum(),
		"sql":             sqlStr,
		"rows":            "-",
		"proc_time":       float64(since.Milliseconds()),
		"current_time":    currentTime,
	}
	switch {
	case err != nil && m.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound)):
		msg["err"] = err
		if rows == -1 {
			Log.TagInfo(traceContext, "_com_mysql_failure", msg)
		} else {
			msg["rows"] = rows
			Log.TagInfo(traceContext, "_com_mysql_failure", msg)
		}
	case since > m.SlowThreshold && m.SlowThreshold != 0 && m.LogLevel >= logger.Warn:
		slowLog := fmt.Sprintf("SLOW SQL>=%v", m.SlowThreshold)
		msg["slowLog"] = slowLog
		if rows == -1 {
			Log.TagInfo(traceContext, "_com_mysql_success", msg)
		} else {
			msg["rows"] = rows
			Log.TagInfo(traceContext, "_com_mysql_success", msg)
		}
	case m.LogLevel == logger.Info:
		if rows == -1 {
			Log.TagInfo(traceContext, "_com_mysql_success", msg)
		} else {
			msg["rows"] = rows
			Log.TagInfo(traceContext, "_com_mysql_success", msg)
		}
	}
}

func CloseDB() error {
	for _, db := range DBMapPool {
		db.Close()
	}
	DBMapPool = make(map[string]*sql.DB)
	GORMMapPool = make(map[string]*gorm.DB)
	return nil
}

func DBPoolLogQuery(trace *TraceContext, sqlDB *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := sqlDB.Query(query, args...)
	end := time.Now()
	if err != nil {
		Log.TagInfo(trace, "_com_mysql_success", map[string]interface{}{
			"sql":       query,
			"bind":      args,
			"proc_time": fmt.Sprintf("%f", end.Sub(start).Seconds()),
		})
	} else {
		Log.TagInfo(trace, "_com_mysql_success", map[string]interface{}{
			"sql":       query,
			"bind":      args,
			"proc_time": fmt.Sprintf("%f", end.Sub(start).Seconds()),
		})
	}
	return rows, err
}
