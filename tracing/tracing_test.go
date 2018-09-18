package tracing_test

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/qeelyn/gin-contrib/tracing"
	"github.com/qeelyn/go-common/logger"
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

	router.Use(tracing.HandleFunc(cnf))
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
	t.Log(tid)
	cnf := map[string]interface{}{"useOpentracing": true}
	go ginServer(cnf,func(g *gin.Context) {
		request := g.Request
		tracer := opentracing.GlobalTracer()
		sid := g.GetString(logger.ContextHeaderName)
		if tid != sid {
			t.Fatal("tid not equal")
			g.Abort()
		}
		var spanCtx opentracing.SpanContext
		if rs,ok := g.Get(tracing.RootSpanContextHeaderName);ok {
			spanCtx =  rs.(opentracing.SpanContext)
		} else {
			spanCtx, _ = tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(request.Header))
		}
		sid2 := spanCtx.(jaeger.SpanContext).TraceID().String()
		if tid != sid2 {
			t.Fatal("tid not equal")
			g.Abort()
		}
		span := tracer.StartSpan("hello-opname", ext.RPCServerOption(spanCtx))
		span.SetTag("hello-tag-key", "hello-tag-value")
		defer span.Finish()
		helloStr := "hello Jaeger"
		span.LogFields(
			log.String("event", "hello-handle"),
			log.String("value", helloStr),
		)
		//ctx := opentracing.ContextWithSpan(g.Background(), span)
		g.Writer.Write([]byte("good"))
	}).ListenAndServe()
	_, err1 := httpClient.Do(httpReq)
	if err1 != nil {
		t.Fatal(err)
	}
}

func TestTracingHandleFuncNoTracer(t *testing.T) {
	tid := tracing.NewTraceId().String()
	cnf := map[string]interface{}{"useOpentracing": false}
	go ginServer(cnf,func(context *gin.Context) {
		sid := context.GetString(logger.ContextHeaderName)
		if tid != sid {
			t.Fatal("tid not equal")
			context.Abort()
		}
		context.Writer.Write([]byte("good"))
	}).ListenAndServe()

	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("GET", "http://localhost:22222", nil)
	httpReq.Header.Set(logger.ContextHeaderName,tid)
	_, err1 := httpClient.Do(httpReq)
	if err1 != nil {
		t.Fatal(err)
	}
}

func TestTracingHandleFuncInTracerSimple(t *testing.T) {
	tracer, closer := newTracer()
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	tid := tracing.NewTraceId().String()
	cnf := map[string]interface{}{"useOpentracing": true}
	go ginServer(cnf,func(context *gin.Context) {
		sid := context.GetString(logger.ContextHeaderName)
		if tid != sid && tid != "" {
			t.Fatal("tid not equal")
			context.Abort()
		}
		context.Writer.Write([]byte("good"))
	}).ListenAndServe()

	httpClient := &http.Client{}
	httpReq, err := http.NewRequest("GET", "http://localhost:22222", nil)
	httpReq.Header.Set(logger.ContextHeaderName,tid)
	_, err1 := httpClient.Do(httpReq)
	if err1 != nil {
		t.Fatal(err)
	}
	tid = ""
	httpClient = &http.Client{}
	httpReq, err = http.NewRequest("GET", "http://localhost:22222", nil)
	_, err1 = httpClient.Do(httpReq)
	if err1 != nil {
		t.Fatal(err)
	}
}


func TestNewTraceId(t *testing.T) {
	_,err := jaeger.TraceIDFromString("106a0d6722fd4b00")
	if err != nil {
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
	}
	tracer, closer, err := cfg.NewTracer(config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}
