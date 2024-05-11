package lib

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
)

type Trace struct {
	TraceId     string
	SpanId      string
	Caller      string
	SrcMethod   string
	HintCode    string
	HintContent string
}

type TraceContext struct {
	Trace
	CSpanId string
}

func SetGinTraceContext(c *gin.Context, trace *TraceContext) error {
	if trace == nil || c == nil {
		return errors.New("context is nil")
	}
	c.Set("trace", trace)
	return nil
}

func GetTraceContext(ctx context.Context) *TraceContext {
	if ginCTX, ok := ctx.(*gin.Context); ok {
		traceIntraceContext, exists := ginCTX.Get("trace")
		if !exists {
			return NewTrace()
		}
		traceContext, ok := traceIntraceContext.(*TraceContext)
		if ok {
			return traceContext
		}
		return NewTrace()
	}

	if ctx2, ok := ctx.(context.Context); ok {
		traceContext, ok := ctx2.Value("trace").(*TraceContext)
		if ok {
			return traceContext
		}
		return NewTrace()
	}

	return NewTrace()
}
