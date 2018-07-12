package tracing

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
)

const GlobalSpan = "globalSpan"
const GlobalTraceId = "globalTraceId"

type JeagerTracer struct {
}

func TracingField(c *gin.Context,z *zap.Logger) zap.Field {
	return zap.String("trace.traceid",c.GetString(GlobalTraceId))
}

func (JeagerTracer) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		sban, _ := opentracing.StartSpanFromContext(c, c.Request.RequestURI)
		tctx := sban.Context().(jaeger.SpanContext)
		tid := tctx.TraceID().String()
		c.Set(GlobalSpan, sban)
		c.Set(GlobalTraceId, tid)
		defer sban.Finish()
		c.Next()
	}
}
