package lib

import (
	"fmt"
	"github.com/m17621679833/nice_base/nlog"
	"strings"
)

const (
	NLTagUndefined     = "_undef"
	NLTagMySqlFailed   = "_com_mysql_failure"
	NLTagRedisFailed   = "_com_redis_failure"
	NLTagMySqlSuccess  = "_com_mysql_success"
	NLTagRedisSuccess  = "_com_redis_success"
	NLTagThriftFailed  = "_com_thrift_failure"
	NLTagThriftSuccess = "_com_thrift_success"
	NLTagHTTPSuccess   = "_com_http_success"
	NLTagHTTPFailed    = "_com_http_failure"
	NLTagTCPFailed     = "_com_tcp_failure"
	NLTagRequestIn     = "_com_request_in"
	NLTagRequestOut    = "_com_request_out"
)
const (
	_nlTag          = "nltag"
	_traceId        = "traceid"
	_spanId         = "spanid"
	_childSpanId    = "cspanid"
	_nlTagBizPrefix = "_com_"
	_nlTagBizUndef  = "_com_undef"
)

var Log *LoggerFaced

type LoggerFaced struct {
}

func (l *LoggerFaced) TagInfo(trace *TraceContext, nltag string, m map[string]interface{}) {
	m[_nlTag] = checkNLTag(nltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	nlog.Info(parseParams(m))
}

func (l *LoggerFaced) TagWarn(trace *TraceContext, nltag string, m map[string]interface{}) {
	m[_nlTag] = checkNLTag(nltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	nlog.Warn(parseParams(m))
}

func (l *LoggerFaced) TagError(trace *TraceContext, nltag string, m map[string]interface{}) {
	m[_nlTag] = checkNLTag(nltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	nlog.Error(parseParams(m))
}

func (l *LoggerFaced) TagTrace(trace *TraceContext, nltag string, m map[string]interface{}) {
	m[_nlTag] = checkNLTag(nltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	nlog.Trace(parseParams(m))
}

func (l *LoggerFaced) TagDebug(trace *TraceContext, nltag string, m map[string]interface{}) {
	m[_nlTag] = checkNLTag(nltag)
	m[_traceId] = trace.TraceId
	m[_childSpanId] = trace.CSpanId
	m[_spanId] = trace.SpanId
	nlog.Debug(parseParams(m))
}

func (l *LoggerFaced) Close() {
	nlog.Close()
}

func checkNLTag(nltag string) string {
	if strings.HasPrefix(nltag, _nlTagBizPrefix) {
		return nltag
	}

	if nltag == NLTagUndefined {
		return nltag
	}
	return nltag
}

// map格式化为string
func parseParams(m map[string]interface{}) string {
	var nltag string = "_undef"
	if _dltag, _have := m["nltag"]; _have {
		if __val, __ok := _dltag.(string); __ok {
			nltag = __val
		}
	}
	for _key, _val := range m {
		if _key == "nltag" {
			continue
		}
		nltag = nltag + "||" + fmt.Sprintf("%v=%+v", _key, _val)
	}
	nltag = strings.Trim(fmt.Sprintf("%q", nltag), "\"")
	return nltag
}

func CreateBizNLTag(tagName string) string {
	if tagName == "" {
		return _nlTagBizUndef
	}
	return _nlTagBizPrefix + tagName
}
