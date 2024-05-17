package test

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/m17621679833/nice_base/lib"
	"testing"
)

func Test_redis(t *testing.T) {
	SetUp()
	c, err := lib.RedisConnFactory("default")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	trace := lib.NewTrace()
	redisKey := "test_key1"
	lib.RedisLogDo(trace, c, "SET", redisKey, "test_dbpool")
	lib.RedisLogDo(trace, c, "expire", "test_key1", 10)

	vint, _ := redis.Int64(lib.RedisLogDo(trace, c, "INCR", "test_incr"))
	fmt.Println(vint)
	// 调用GET
	v, err := redis.String(lib.RedisLogDo(trace, c, "GET", redisKey))
	fmt.Println(v)
	if v != "test_dbpool" || err != nil {
		t.Fatal(err)
	}

	// 使用RedisConfDo调用GET
	v2, err := redis.String(lib.RedisConfDo(trace, "default", "GET", redisKey))
	fmt.Println(v2)
	fmt.Println(err)
	if v != "test_dbpool" || err != nil {
		t.Fatal("test redis get fatal!")
	}
}
