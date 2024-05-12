package test

import (
	"github.com/gin-gonic/gin"
	"github.com/m17621679833/nice_base/lib"
	"testing"
)

func TestGetTraceContext(t *testing.T) {
	trace := &lib.TraceContext{
		Trace: lib.Trace{
			TraceId: "dc traceId",
		},
		CSpanId: "cs span",
	}

	context := &gin.Context{}
	lib.SetGinTraceContext(context, trace)
	traceContext := lib.GetTraceContext(context)
	t.Log(traceContext)
}
