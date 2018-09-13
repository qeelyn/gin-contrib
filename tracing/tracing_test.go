package tracing_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/qeelyn/gin-contrib/tracing"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"io"
	"net/http"
	"testing"
)

func ginServer(cnf map[string]interface{},fun func(context *gin.Context)) *http.Server {
	router := gin.New()
	httpServer := &http.Server{
		Addr:    ":22222",
		Handler: router,
	}

	router.Use(tracing.TracingHandleFunc(cnf))
	router.GET("/", fun)
	return httpServer
}

func TestTracingHandleFunc(t *testing.T) {
	tracer, closer := newTracer()
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)


	span := opentracing.StartSpan("test")
	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("GET", "http://localhost:22222", nil)
	if err != nil {
		t.Fatal(err)
	}
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(httpReq.Header))
	span.Finish()

	tid := span.Context().(jaeger.SpanContext).TraceID().String()
	cnf := map[string]interface{}{"useOpentracing": true}
	go ginServer(cnf,func(context *gin.Context) {
		request := context.Request
		tracer := opentracing.GlobalTracer()
		sid := context.GetString(tracing.GlobalTraceId)
		if tid != sid {
			t.Fatal("tid not equal")
			context.Abort()
		}
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header))
		sid2 := spanCtx.(jaeger.SpanContext).TraceID().String()
		if tid != sid2 {
			t.Fatal("tid not equal")
			context.Abort()
		}
		span := tracer.StartSpan("hello-opname", ext.RPCServerOption(spanCtx))
		span.SetTag("hello-tag-key", "hello-tag-value")
		defer span.Finish()
		helloStr := "hello Jaeger"
		span.LogFields(
			log.String("event", "hello-handle"),
			log.String("value", helloStr),
		)
		//ctx := opentracing.ContextWithSpan(context.Background(), span)
		context.Writer.Write([]byte("good"))
	}).ListenAndServe()
	_, err1 := httpClient.Do(httpReq)
	if err1 != nil {
		t.Fatal(err)
	}

}

func TestTracingHandleFuncNoTracer(t *testing.T) {
	tid := uuid.New().String()
	cnf := map[string]interface{}{"useOpentracing": false}
	go ginServer(cnf,func(context *gin.Context) {
		sid := context.GetString(tracing.GlobalTraceId)
		if tid != sid {
			t.Fatal("tid not equal")
			context.Abort()
		}
		context.Writer.Write([]byte("good"))
	}).ListenAndServe()

	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("GET", "http://localhost:22222", nil)
	httpReq.Header.Set(tracing.GlobalTraceId,tid)
	_, err1 := httpClient.Do(httpReq)
	if err1 != nil {
		t.Fatal(err)
	}
}

func newTracer() (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		ServiceName: "test",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
		Headers: &jaeger.HeadersConfig{
			TraceContextHeaderName: tracing.GlobalTraceId,
		},
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}
