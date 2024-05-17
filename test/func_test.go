package test

import (
	"bytes"
	"fmt"
	"github.com/m17621679833/nice_base/lib"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"
)

var (
	addr                 = "127.0.0.1:6111"
	initOnce   sync.Once = sync.Once{}
	serverOnce sync.Once = sync.Once{}
)

type HttpConf struct {
	ServerAddr     string   `mapstructure:"server_addr"`
	ReadTimeout    int      `mapstructure:"read_timeout"`
	WriteTimeout   int      `mapstructure:"write_timeout"`
	MaxHeaderBytes int      `mapstructure:"max_header_bytes"`
	AllowHost      []string `mapstructure:"allow_host"`
}

func Test_GetConfEnv(t *testing.T) {
	SetUp()
	fmt.Println(lib.GetConfEnv())
}

func Test_ParseLocalConfig(t *testing.T) {
	SetUp()
	conf := HttpConf{}
	err := lib.ParseLocalConfig("test.toml", conf)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(conf)
	//TearDown()
}
func SetUp() {
	initOnce.Do(func() {
		err := lib.InitModule("../conf/dev/", []string{"base", "mysql", "redis"})
		if err != nil {
			log.Fatal(err)
		}
	})
}

func TearDown() {
	lib.Destroy()
}

func TestFunc(t *testing.T) {
	InitTestServer()
}

func InitTestServer() {
	serverOnce.Do(func() {
		http.HandleFunc("/postjson", func(writer http.ResponseWriter, request *http.Request) {
			data, err := io.ReadAll(request.Body)
			if err != nil {
				log.Fatal(err)
			}
			request.Body = io.NopCloser(bytes.NewBuffer(data))
			writer.Write(data)
		})
	})

	http.HandleFunc("/get", func(write http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		cityId := request.FormValue("city_id")
		write.Write([]byte(cityId))
	})
	http.HandleFunc("/post", func(writer http.ResponseWriter, request *http.Request) {
		request.ParseForm()
		value := request.FormValue("city_id")
		writer.Write([]byte(value))
	})

	go func() {
		log.Println("ListenAndServe ", addr)
		http.ListenAndServe(addr, nil)

	}()
	time.Sleep(time.Second * 1)
}

func TestGet(t *testing.T) {
	InitTestServer()
	a := url.Values{
		"city_id": {"12"},
	}
	url := "http://" + addr + "/get"
	_, i, err := lib.HttpGet(lib.NewTrace(), url, a, 1000, nil)
	fmt.Println("city_id=" + string(i))
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestJson(t *testing.T) {
	InitTestServer()
	jsonStr := "{\"source\":\"control\",\"cityId\":\"12\",\"trailNum\":10,\"dayTime\":\"2018-11-21 16:08:00\",\"limit\":2,\"andOperations\":{\"cityId\":\"eq\",\"trailNum\":\"gt\",\"dayTime\":\"eq\"}}"
	url := "http://" + addr + "/postjson"
	_, i, err := lib.HttpJson(lib.NewTrace(), url, jsonStr, 1000, nil)
	fmt.Println(string(i))
	if err != nil {
		fmt.Println(err.Error())
	}
}
