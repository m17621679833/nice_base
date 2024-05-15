package lib

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func Init(configPath string) error {
	return InitModule(configPath, []string{"base", "mysql", "redis"})
}

func InitModule(configPath string, modules []string) error {
	conf := flag.String("config", configPath, "input config file like ./conf/dev/")
	flag.Parse()
	if *conf == "" {
		flag.Usage()
		os.Exit(1)
	}

	log.Println("-------------------------------------------------")
	log.Printf("[INFO] config=%s\n", *conf)
	log.Printf("[INFO]%s\n", "start loading resources.")
	ips := GetLocalIPs()
	if len(ips) > 0 {
		LocalIP = ips[0]
	}
	if err := ParseConfigPath(*conf); err != nil {
		return err
	}

	if err := InitViperConf(); err != nil {
		return err
	}

	if InArrayString("base", modules) {
		if err := InitBaseConf(GetConfPath("base")); err != nil {
			fmt.Printf("[ERROR] %s%s\n", time.Now().Format(TimeFormat), " Init base conf:"+err.Error())
		}
	}

	if InArrayString("redis", modules) {
		if err := InitRedisConf(GetConfPath("redis_map")); err != nil {
			fmt.Printf("[ERROR] %s%s\n", time.Now().Format(TimeFormat), " Init redis conf:"+err.Error())
		}
	}

	if InArrayString("mysql", modules) {
		if err := InitDBPool(GetConfPath("mysql_map")); err != nil {
			fmt.Printf("[ERROR]%s%s\n", time.Now().Format(TimeFormat), "Init mysql conf:"+err.Error())
		}
	}

	if location, err := time.LoadLocation(ConfBase.TimeLocation); err != nil {
		return err
	} else {
		TimeLocation = location
	}

	log.Printf("[INFO] %s\n", " success loading resources.")
	log.Println("------------------------------------------------------------------------")
	return nil
}

func GetConfPath(fileName string) string {
	return ConfEnvPath + "/" + fileName + ".toml"
}

/*
./conf/dev/
*/
func ParseConfigPath(config string) error {
	path := strings.Split(config, "/")
	ConfEnvPath = strings.Join(path[:len(path)-1], "/")
	ConfEnv = path[len(path)-2]
	return nil
}

func NewTrace() *TraceContext {
	trace := &TraceContext{}
	trace.TraceId = GetTraceId()
	trace.SpanId = NewSpanId()
	return trace
}

func NewSpanId() string {
	timestamp := uint32(time.Now().Unix())
	ipToLong := binary.BigEndian.Uint32(LocalIP.To4())
	b := bytes.Buffer{}
	b.WriteString(fmt.Sprintf("%08x", ipToLong^timestamp))
	b.WriteString(fmt.Sprintf("%08x", rand.Int31()))
	return b.String()
}

func GetTraceId() (traceId string) {
	return calTraceId(LocalIP.String())
}

func calTraceId(ip string) (traceId string) {
	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getgid()

	b := bytes.Buffer{}
	netIp := net.ParseIP(ip)
	if netIp == nil {
		b.WriteString("00000000")
	} else {
		b.WriteString(hex.EncodeToString(netIp.To4()))
	}
	b.WriteString(fmt.Sprintf("%08x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%04x", rand.Int31n(1<<24)))
	b.WriteString("b0")

	return b.String()
}

func GetLocalIPs() (ips []net.IP) {
	interfaceAddrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, addr := range interfaceAddrs {
		ipNet, isValidIpNet := addr.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}
	return ips
}

func InArrayString(s string, arr []string) bool {
	for _, i := range arr {
		if i == s {
			return true
		}
	}
	return false
}

func Substr(str string, start int64, end int64) string {
	length := int64(len(str))
	if start < 0 || start > length {
		return ""
	}
	if end > length {
		end = length
	}
	return string(str[start:end])
}
