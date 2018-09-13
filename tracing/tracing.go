package tracing

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
)

var (
	GlobalTraceId = "trace.traceid"
)

func TracingField(c *gin.Context, z *zap.Logger) zap.Field {
	return zap.String(GlobalTraceId, c.GetString(GlobalTraceId))
}

// default receive http header `trace.traceid` for tracing
// use 'useOpentracing': true to enable JaegerTracer
func TracingHandleFunc(config map[string]interface{}) gin.HandlerFunc {
	if tid, ok := config[GlobalTraceId]; ok {
		GlobalTraceId = tid.(string)
	}
	useOpentracing, _ := config["useOpentracing"].(bool)
	return func(c *gin.Context) {
		if (useOpentracing) {
			ctx, err := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(c.Request.Header))

			if err != nil {
				c.Error(err)
				return
			} else if ctx == nil {
				sban, _ := opentracing.StartSpanFromContext(c, c.Request.RequestURI)
				ctx = sban.Context()
				defer sban.Finish()
			}
			jaegerCtx := ctx.(jaeger.SpanContext)
			tid := jaegerCtx.TraceID().String()
			c.Set(GlobalTraceId, tid)
		} else {
			ctid := c.Request.Header.Get(GlobalTraceId)
			if ctid == "" {
				ctid = uuid.New().String()
			}
			c.Set(GlobalTraceId, ctid)
		}
		c.Next()
	}
}
