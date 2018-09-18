module github.com/qeelyn/gin-contrib

replace (
	golang.org/x/net => github.com/golang/net v0.0.0-20180811021610-c39426892332
	golang.org/x/sys => github.com/golang/sys v0.0.0-20180810173357-98c5dad5d1a0
	golang.org/x/text => github.com/golang/text v0.3.0
	google.golang.org/appengine => gitee.com/githubmirror/appengine v1.1.0
	google.golang.org/genproto => gitee.com/githubmirror/go-genproto v0.0.0-20180808183934-383e8b2c3b9e
	google.golang.org/grpc => gitee.com/githubmirror/grpc-go v1.14.0
)

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/sse v0.0.0-20170109093832-22d885f9ecc7 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/google/uuid v1.0.0 // indirect
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/opentracing/opentracing-go v1.0.2
	github.com/qeelyn/go-common v0.0.0-20180918163346-2e73381a4f74
	github.com/uber/jaeger-client-go v2.14.0+incompatible
	go.uber.org/zap v1.9.1
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/yaml.v2 v2.2.1
)
