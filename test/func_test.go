package test

import (
	"bytes"
	"fmt"
	"github.com/m17621679833/nice_base/lib"
	"io"
	"log"
	"net/http"
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

func SetUp() {
	initOnce.Do(func() {
		err := lib.InitModule("../conf/dev/", []string{"base", "mysql", "redis"})
		if err != nil {
			log.Fatal(err)
		}
	})
}

func TearDown() {

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
			writer.Write([]byte(data))
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
	time.Sleep(time.Second * 200)
}
