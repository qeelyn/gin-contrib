package tracing

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"github.com/uber/jaeger-client-go"
)

const GlobalSpan = "trace.spanid"
const GlobalTraceId = "trace.traceid"

type JeagerTracer struct {
}

func TracingField(c *gin.Context,z *zap.Logger) zap.Field {
	return zap.String(GlobalTraceId,c.GetString(GlobalTraceId))
}

func (JeagerTracer) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		sban, _ := opentracing.StartSpanFromContext(c, c.Request.RequestURI)
		if sban != nil {
			if tctx,ok := sban.Context().(jaeger.SpanContext);ok {
				tid := tctx.TraceID().String()
				c.Set(GlobalSpan, sban)
				c.Set(GlobalTraceId, tid)
				defer sban.Finish()
			}
		}
		c.Next()
	}
}
