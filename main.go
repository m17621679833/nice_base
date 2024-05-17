package main

import (
	"fmt"
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
	err := lib.GORMDefaultPool.Table("agv_identity_info").Find(idf).Error
	if err != nil {

	}
	fmt.Println(idf)
	lib.Log.TagInfo(lib.NewTrace(), lib.NLTagUndefined, map[string]interface{}{
		"message": "todo sth",
	})
	time.Sleep(time.Second)
}
