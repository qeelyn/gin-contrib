package tracing

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/qeelyn/go-common/logger"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/utils"
	"time"
)

var (
	// the header key is Qeelyn-Traceid,match the opentracing's tag id rule
	RootSpanContextHeaderName = "qeelyn-spanid"
	// accept http request,not use opentracing
	HttpHeaderName = "qeelyn-tracing-id"
)

const (
	// opentracing log key is trace.traceid
	LoggerFieldKey = "traceid"
)

// default receive http header `trace.traceid` for tracing
// use 'useOpentracing': true to enable JaegerTracer
func HandleFunc(config map[string]interface{}) gin.HandlerFunc {
	useOpentracing, _ := config["useOpentracing"].(bool)
	return func(c *gin.Context) {
		var (
			tid jaeger.TraceID
			err error
		)
		gid := c.Request.Header.Get(HttpHeaderName)
		if useOpentracing && gid == "" {
			ctx, _ := opentracing.GlobalTracer().Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(c.Request.Header))
			jaegerCtx := ctx.(jaeger.SpanContext)
			if jaegerCtx.IsValid() {
				c.Set(RootSpanContextHeaderName, jaegerCtx)
				tid = jaegerCtx.TraceID()
			} else {
				tid = NewTraceId()
			}
		} else if gid != "" {
			if tid, err = jaeger.TraceIDFromString(gid); err != nil {
				tid = NewTraceId()
			}
		} else {
			tid = NewTraceId()
		}
		c.Set(logger.ContextHeaderName, tid.String())
		c.Next()
	}
}

func NewTraceId() jaeger.TraceID {
	traceID := jaeger.TraceID{}
	traceID.Low = randomID()
	if traceID.Low == 0 {
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

func SpanFromContext(g *gin.Context) (opentracing.SpanContext, error) {
	span, ok := g.Get(RootSpanContextHeaderName)
	if !ok {
		return nil, errors.New("span not found")
	}
	if sctx, ok := span.(opentracing.SpanContext); ok {
		return sctx, nil
	} else {
		return nil, errors.New("span is not SpanContext")
	}
}
