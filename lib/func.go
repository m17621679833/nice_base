package lib

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/m17621679833/nice_base/nlog"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
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

func Destroy() {
	log.Println("------------------------------------------------------------------------")
	log.Printf("[INFO] %s\n", " start destroy resources.")
	CloseDB()
	nlog.Close()
	log.Printf("[INFO] %s\n", " success destroy resources.")
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

func HttpGet(trace *TraceContext, urlString string, urlParams url.Values, msTimeout int, header http.Header) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{Timeout: time.Duration(msTimeout) * time.Millisecond}
	urlString = AddGetDataToUrl(urlString, urlParams)
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		Log.TagWarn(trace, NLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	if len(header) > 0 {
		req.Header = header
	}
	req = addTrace2Header(req, trace)
	response, err := client.Do(req)
	if err != nil {
		Log.TagWarn(trace, NLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		Log.TagWarn(trace, NLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "GET",
			"args":      urlParams,
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	Log.TagInfo(trace, NLTagHTTPSuccess, map[string]interface{}{
		"url":       urlString,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "GET",
		"args":      urlParams,
		"result":    Substr(string(data), 0, 1024),
	})
	return response, data, nil
}

func HttpPost(trace *TraceContext, urlString string, urlParams url.Values, msTimeout int, header http.Header, contextType string) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{
		Timeout: time.Duration(msTimeout) * time.Millisecond,
	}
	if contextType == "" {
		contextType = "application/x-www-form-urlencoded"
	}
	urlParamEncode := urlParams.Encode()
	request, err := http.NewRequest("POST", urlString, strings.NewReader(urlParamEncode))
	if len(request.Header) > 0 {
		request.Header = header
	}
	request = addTrace2Header(request, trace)
	request.Header.Set("Content-Type", contextType)
	resp, err := client.Do(request)
	if err != nil {
		Log.TagWarn(trace, NLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(urlParamEncode, 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		Log.TagWarn(trace, NLTagHTTPFailed, map[string]interface{}{
			"url":       urlString,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(urlParamEncode, 0, 1024),
			"err":       err.Error(),
		})
	}
	Log.TagInfo(trace, NLTagHTTPSuccess, map[string]interface{}{
		"url":       urlString,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "POST",
		"args":      Substr(urlParamEncode, 0, 1024),
		"result":    Substr(string(body), 0, 1024),
	})
	return resp, body, nil
}

func HttpJson(trace *TraceContext, urlStrings string, jsonContent string, msTimeout int, header http.Header) (*http.Response, []byte, error) {
	startTime := time.Now().UnixNano()
	client := http.Client{
		Timeout: time.Duration(msTimeout) * time.Millisecond,
	}
	req, err := http.NewRequest("POST", urlStrings, strings.NewReader(jsonContent))
	if len(header) > 0 {
		req.Header = header
	}
	req = addTrace2Header(req, trace)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		Log.TagWarn(trace, NLTagHTTPFailed, map[string]interface{}{
			"url":       urlStrings,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(jsonContent, 0, 1924),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	defer resp.Body.Close()
	ret, err := io.ReadAll(resp.Body)
	if err != nil {
		Log.TagWarn(trace, NLTagHTTPFailed, map[string]interface{}{
			"url":       urlStrings,
			"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
			"method":    "POST",
			"args":      Substr(jsonContent, 0, 1024),
			"err":       err.Error(),
		})
		return nil, nil, err
	}
	Log.TagInfo(trace, NLTagHTTPSuccess, map[string]interface{}{
		"url":       urlStrings,
		"proc_time": float32(time.Now().UnixNano()-startTime) / 1.0e9,
		"method":    "POST",
		"args":      Substr(jsonContent, 0, 1024),
		"result":    Substr(string(ret), 0, 1024),
	})
	return resp, ret, nil

}

func addTrace2Header(req *http.Request, trace *TraceContext) *http.Request {
	traceId := trace.TraceId
	cSpanId := NewSpanId()
	if traceId != "" {
		req.Header.Set("nice-header-rid", traceId)
	}
	if cSpanId != "" {
		req.Header.Set("nice-header-spanid", cSpanId)
	}
	trace.SpanId = cSpanId
	return req
}

func AddGetDataToUrl(urlString string, params url.Values) string {
	if strings.Contains(urlString, "?") {
		urlString += "&"
	} else {
		urlString += "?"
	}
	return fmt.Sprintf("%s%s", urlString, params.Encode())
}
