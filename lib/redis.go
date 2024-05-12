package lib

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"time"
)

func RedisConnFactory(name string) (redis.Conn, error) {
	if ConfRedisMap != nil && ConfRedisMap.List != nil {
		for confName, conf := range ConfRedisMap.List {
			if name == confName {
				randHost := conf.ProxyList[rand.Intn(len(conf.ProxyList))]
				if conf.ConnTimeout == 0 {
					conf.ConnTimeout = 50
				}
				if conf.ReadTimeout == 0 {
					conf.ReadTimeout = 100
				}
				if conf.WriteTimeout == 0 {
					conf.WriteTimeout = 100
				}

				c, err := redis.Dial("tcp", randHost, redis.DialConnectTimeout(time.Duration(conf.ConnTimeout)*time.Millisecond),
					redis.DialReadTimeout(time.Duration(conf.ReadTimeout)*time.Millisecond),
					redis.DialWriteTimeout(time.Duration(conf.WriteTimeout)*time.Millisecond))
				if err != nil {
					return nil, err
				}
				if conf.Password != "" {
					if _, err := c.Do("AUTH", conf.Password); err != nil {
						c.Close()
						return nil, err
					}
				}
				if conf.Db != 0 {
					if _, err := c.Do("SELECT", conf.Db); err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, nil
			}
		}
	}
	return nil, errors.New("create redis conn fail")
}

func RedisLogDo(trace *TraceContext, c redis.Conn, commandName string, args ...interface{}) {
	start := time.Now()
	reply, err := c.Do(commandName, args...)
	end := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    commandName,
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", end.Sub(start).Seconds()),
		})
	} else {
		replyStr, _ := redis.String(reply, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":    commandName,
			"bind":      args,
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", end.Sub(start).Seconds()),
		})
	}
}
func RedisConfDo(trace *TraceContext, name string, commandName string, args ...interface{}) (interface{}, error) {
	c, err := RedisConnFactory(name)
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method": commandName,
			"err":    errors.New("RedisConnFactory_error:" + name),
			"bind":   args,
		})
		return nil, err
	}
	defer c.Close()

	startExecTime := time.Now()
	reply, err := c.Do(commandName, args...)
	endExecTime := time.Now()
	if err != nil {
		Log.TagError(trace, "_com_redis_failure", map[string]interface{}{
			"method":    commandName,
			"err":       err,
			"bind":      args,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	} else {
		replyStr, _ := redis.String(reply, nil)
		Log.TagInfo(trace, "_com_redis_success", map[string]interface{}{
			"method":    commandName,
			"bind":      args,
			"reply":     replyStr,
			"proc_time": fmt.Sprintf("%fs", endExecTime.Sub(startExecTime).Seconds()),
		})
	}
	return reply, err
}
