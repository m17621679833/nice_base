package main

import (
	"github.com/m17621679833/nice_base/lib"
	"log"
	"time"
)

func main() {
	if err := lib.InitModule("./conf/dev/", []string{"base", "mysql", "redis"}); err != nil {
		log.Fatal(err)
	}
	defer lib.Destroy()
	type IdentifyInfo struct {
		Id int `json:"id" gorm:"primary_key" description:"自增主键"`
	}
	idf := &IdentifyInfo{}
	lib.GORMDefaultPool.Table("agv_identity_info").Find(idf)

	time.Sleep(time.Second)
}
