package errorhandle_test

import (
	"github.com/gin-gonic/gin"
	"github.com/qeelyn/gin-contrib/errorhandle"
	"github.com/qeelyn/go-common/logger"
	"testing"
)

func TestErrorHandle(t *testing.T) {
	config := map[string]interface{}{}
	efunc := errorhandle.ErrorHandle(config,logger.NewLogger())

	efunc(&gin.Context{})
}