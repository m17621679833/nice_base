package test

import (
	"context"
	"fmt"
	"github.com/m17621679833/nice_base/lib"
	"gorm.io/gorm"
	"testing"
	"time"
)

type Test1 struct {
	Id        int64     `json:"id" gorm:"primary_key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func (f *Test1) Table() string {
	return "test1"
}

func (f *Test1) DB() *gorm.DB {
	return lib.GORMDefaultPool
}

var (
	createTableSQL = "CREATE TABLE `test1` (`id` int(12) unsigned NOT NULL AUTO_INCREMENT" +
		" COMMENT '自增id',`name` varchar(255) NOT NULL DEFAULT '' COMMENT '姓名'," +
		"`created_at` datetime NOT NULL,PRIMARY KEY (`id`)) ENGINE=InnoDB " +
		"DEFAULT CHARSET=utf8"
	insertSQL    = "INSERT INTO `test1` (`id`, `name`, `created_at`) VALUES (NULL, '111', '2018-08-29 11:01:43');"
	dropTableSQL = "DROP TABLE `test1`"
	beginSQL     = "start transaction;"
	commitSQL    = "commit;"
	rollbackSQL  = "rollback;"
)

func Test_DBPool(t *testing.T) {
	SetUp()

	dbPool, err := lib.GetDBPool("default")
	if err != nil {
		t.Fatal(err)
	}
	trace := lib.NewTrace()
	if _, err = lib.DBPoolLogQuery(trace, dbPool, beginSQL); err != nil {
		t.Fatal(err)
	}

	if _, err := lib.DBPoolLogQuery(trace, dbPool, createTableSQL); err != nil {
		lib.DBPoolLogQuery(trace, dbPool, rollbackSQL)
		t.Fatal(err)
	}

	if _, err := lib.DBPoolLogQuery(trace, dbPool, insertSQL); err != nil {
		lib.DBPoolLogQuery(trace, dbPool, rollbackSQL)
		t.Fatal(err)
	}

	current_id := 0
	table_name := "test1"
	fmt.Println("begin read table", table_name, "")
	fmt.Println("-----------------------------------------------------")
	fmt.Printf("%6s | %6s\n\n", "id", "created_at")
	for {
		rows, err := lib.DBPoolLogQuery(trace, dbPool, "SELECT id,created_at FROM test1 WHERE id>? order by id asc", current_id)
		defer rows.Close()
		row_len := 0
		if err != nil {
			lib.DBPoolLogQuery(trace, dbPool, "rollback;")
			t.Fatal(err)
		}
		for rows.Next() {
			var create_time string
			err := rows.Scan(&current_id, &create_time)
			if err != nil {
				lib.DBPoolLogQuery(trace, dbPool, "rollback;")
				t.Fatal(err)
			}
			fmt.Printf("%6d | %6s\n", current_id, create_time)
			row_len++
		}
		if row_len == 0 {
			break
		}
	}
	fmt.Println("-----------------------------------------------------")
	fmt.Println("finish read table", table_name, "")
	//删除表
	if _, err := lib.DBPoolLogQuery(trace, dbPool, dropTableSQL); err != nil {
		lib.DBPoolLogQuery(trace, dbPool, rollbackSQL)
		t.Fatal(err)
	}

	//提交事务
	lib.DBPoolLogQuery(trace, dbPool, commitSQL)

}

func Test_GORM(t *testing.T) {
	SetUp()

	dbpool, err := lib.GetGormPool("default")
	if err != nil {
		t.Fatal(err)
	}
	db := dbpool.Begin()
	traceCTX := lib.NewTrace()
	ctx := context.Background()
	lib.SetTraceContext(ctx, traceCTX)
	db = db.WithContext(ctx)

	if err := db.Exec(createTableSQL).Error; err != nil {
		db.Rollback()
		t.Fatal(err)
	}

	t1 := &Test1{Name: "test_name", CreatedAt: time.Now()}
	err = db.Save(t1).Error
	if err != err {
		db.Rollback()
		t.Fatal(err)
	}

	list := []Test1{}
	if err := db.Where("name=?", "test_name").Find(&list).Error; err != nil {
		db.Rollback()
		t.Fatal(err)
	}
	fmt.Println(list)

	if err := db.Exec(dropTableSQL).Error; err != nil {
		db.Rollback()
		t.Fatal(err)
	}
	db.Commit()

}
