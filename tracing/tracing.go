package tracing

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/utils"
	"go.uber.org/zap"
	"time"
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
				c.Next()
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
			var (
				tid jaeger.TraceID
				err error
			)
			gid := c.Request.Header.Get(GlobalTraceId)
			if gid == "" {
				tid = NewTraceId()
			} else {
				if tid,err = jaeger.TraceIDFromString(gid);err != nil {
					c.Error(err)
					tid = NewTraceId()
				}
			}
			c.Set(GlobalTraceId, tid.String())
		}
		c.Next()
	}
}

func NewTraceId() jaeger.TraceID {
	traceID := jaeger.TraceID{}
	traceID.Low = randomID()
	if traceID.Low ==0 {
		traceID.Low = randomID()
	}
	return traceID
}

// randomID generates a random trace/span ID, using tracer.random() generator.
// It never returns 0.
func randomID() uint64 {
	rng := utils.NewRand(time.Now().UnixNano())
	return uint64(rng.Int63())
}