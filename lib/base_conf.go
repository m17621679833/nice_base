package lib

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/m17621679833/nice_base/nlog"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var ConfBase *BaseConf
var DBMapPool map[string]*sql.DB
var GORMMapPool map[string]*gorm.DB
var DBDefaultPool *sql.DB
var GORMDefaultPool *gorm.DB
var ConfRedis *RedisConf
var ConfRedisMap *RedisConfMap
var ViperConfMap map[string]*viper.Viper
var ConfEnvPath string
var ConfEnv string

type LogConfFileWriter struct {
	On              bool   `mapstructure:"on"`
	LogPath         string `mapstructure:"log_path"`
	RotateLogPath   string `mapstructure:"rotate_log_path"`
	WfLogPath       string `mapstructure:"wf_log_path"`
	RotateWfLogPath string `mapstructure:"rotate_wf_log_path"`
}

type LogConfConsoleWriter struct {
	On    bool `mapstructure:"on"`
	Color bool `mapstructure:"color"`
}

type LogConfig struct {
	Level string               `mapstructure:"log_level"`
	FW    LogConfFileWriter    `mapstructure:"file_writer"`
	CW    LogConfConsoleWriter `mapstructure:"console_writer"`
}

type BaseConf struct {
	DebugMode    string    `mapstructure:"debug_mode"`
	TimeLocation string    `mapstructure:"time_location"`
	Log          LogConfig `mapstructure:"log"`
	Base         struct {
		DebugMode    string `mapstructure:"debug_mode"`
		TimeLocation string `mapstructure:"time_location"`
	} `mapstructure:"base"`
}

type MysqlConfMap struct {
	List map[string]*MysqlConf `mapstructure:"list"`
}

type MysqlConf struct {
	DriverName      string `mapstructure:"driver_name"`
	DataSourceName  string `mapstructure:"data_source_name"`
	MaxOpenConn     int    `mapstructure:"max_open_conn"`
	MaxIdleConn     int    `mapstructure:"max_idle_conn"`
	MaxConnLifeTime int    `mapstructure:"max_conn_life_time"`
}

type RedisConfMap struct {
	List map[string]*RedisConf `mapstructure:"list"`
}

type RedisConf struct {
	ProxyList    []string `mapstructure:"proxy_list"`
	Password     string   `mapstructure:"password"`
	Db           int      `mapstructure:"db"`
	ConnTimeout  int      `mapstructure:"conn_timeout"`
	ReadTimeout  int      `mapstructure:"read_timeout"`
	WriteTimeout int      `mapstructure:"write_timeout"`
}

func GetBaseConf() *BaseConf {
	return ConfBase
}

func GetStringConf(key string) string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return ""
	}
	v, ok := ViperConfMap[keys[0]]
	if !ok {
		return ""
	}
	confString := v.GetString(strings.Join(keys[1:len(keys)], "."))
	return confString
}

func GetStringMapConf(key string) map[string]interface{} {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetStringMap(strings.Join(keys[1:len(keys)], "."))
	return conf
}

func GetConf(key string) interface{} {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return false
	}
	v := ViperConfMap[keys[0]]
	conf := v.Get(strings.Join(keys[1:len(keys)], "."))
	return conf
}

func GetBoolConf(key string) bool {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return false
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetBool(strings.Join(keys[1:len(keys)], "."))
	return conf
}

// 获取get配置信息
func GetFloat64Conf(key string) float64 {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetFloat64(strings.Join(keys[1:len(keys)], "."))
	return conf
}

// 获取get配置信息
func GetIntConf(key string) int {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetInt(strings.Join(keys[1:len(keys)], "."))
	return conf
}

// 获取get配置信息
func GetStringMapStringConf(key string) map[string]string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetStringMapString(strings.Join(keys[1:len(keys)], "."))
	return conf
}

// 获取get配置信息
func GetStringSliceConf(key string) []string {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return nil
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetStringSlice(strings.Join(keys[1:len(keys)], "."))
	return conf
}

// 获取get配置信息
func GetTimeConf(key string) time.Time {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return time.Now()
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetTime(strings.Join(keys[1:len(keys)], "."))
	return conf
}

// 获取时间阶段长度
func GetDurationConf(key string) time.Duration {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return 0
	}
	v := ViperConfMap[keys[0]]
	conf := v.GetDuration(strings.Join(keys[1:len(keys)], "."))
	return conf
}

// 是否设置了key
func IsSetConf(key string) bool {
	keys := strings.Split(key, ".")
	if len(keys) < 2 {
		return false
	}
	v := ViperConfMap[keys[0]]
	conf := v.IsSet(strings.Join(keys[1:len(keys)], "."))
	return conf
}

func InitBaseConf(path string) error {
	confBase := &BaseConf{}
	err := ParseConfig(path, confBase)
	if err != nil {
		return err
	}
	if ConfBase.DebugMode == "" {
		if confBase.Base.DebugMode != "" {
			confBase.DebugMode = confBase.Base.DebugMode
		} else {
			confBase.DebugMode = "debug"
		}
	}
	if confBase.TimeLocation == "" {
		if confBase.Base.TimeLocation != "" {
			confBase.TimeLocation = confBase.Base.TimeLocation
		} else {
			confBase.TimeLocation = "Asia/Chongqing"
		}
	}
	if confBase.Log.Level == "" {
		confBase.Log.Level = "trace"
	}
	logConfig := &nlog.LogConfig{
		LogLevel: confBase.Log.Level,
		FileWriter: nlog.FileWriterConf{
			On:              confBase.Log.FW.On,
			LogPath:         confBase.Log.FW.LogPath,
			RotateLogPath:   confBase.Log.FW.RotateLogPath,
			WfLogPath:       confBase.Log.FW.WfLogPath,
			RotateWfLogPath: confBase.Log.FW.RotateWfLogPath,
		},
		ConsoleWriter: nlog.ConsoleWriterConf{
			On:    confBase.Log.CW.On,
			Color: confBase.Log.CW.Color,
		},
	}
	err = nlog.SetupDefaultLogWithConf(logConfig)
	if err != nil {
		panic(err)
	}
	nlog.SetLayout("2024-05-10T15:21:23.000")
	return nil
}

func InitRedisConf(path string) error {
	redisConf := &RedisConfMap{}
	err := ParseConfig(path, redisConf)
	if err != nil {
		return err
	}
	ConfRedisMap = redisConf
	return nil
}

func InitViperConf() error {
	file, err := os.Open(ConfEnvPath + "/")
	if err != nil {
		return nil
	}
	fileList, err := file.ReadDir(1024)
	if err != nil {
		return err
	}
	for _, innerFile := range fileList {
		data, err := ioutil.ReadFile(ConfEnvPath + "/" + innerFile.Name())
		if err != nil {
			return err
		}
		v := viper.New()
		v.SetConfigType("toml")
		v.ReadConfig(bytes.NewBuffer(data))
		pathArray := strings.Split(innerFile.Name(), ".")
		if ViperConfMap == nil {
			ViperConfMap = make(map[string]*viper.Viper)
		}
		ViperConfMap[pathArray[0]] = v
	}
	return nil
}

func ParseConfig(path string, conf interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config %v fail,%v", path, err)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read config error:%v", err)
	}
	v := viper.New()
	v.SetConfigType("toml")
	v.ReadConfig(bytes.NewBuffer(data))
	if err = v.Unmarshal(conf); err != nil {
		return fmt.Errorf("unmarshal %v error %v\n", string(data), err)
	}
	return nil
}
