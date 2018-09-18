package errorhandle

import (
	"github.com/gin-gonic/gin"
	"github.com/qeelyn/gin-contrib/tracing"
	"github.com/qeelyn/go-common/errors"
	"github.com/qeelyn/go-common/logger"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var ErrMessage *errors.ErrorMessage

func ErrorHandle(config map[string]interface{}, logger *logger.Logger) gin.HandlerFunc {
	bytes, err := ioutil.ReadFile(config["error-template"].(string))
	if err != nil {
		panic(err)
	}
	templates := map[string]errors.ErrorTemplate{}
	yaml.Unmarshal(bytes, &templates)
	ErrMessage = errors.NewErrorMessage(&templates)

	return func(c *gin.Context) {
		c.Next()
		if l := len(c.Errors); l > 0 {
			var (
				errArray = make([]*errors.ErrorDescription, l)
				status   = 500
			)

			for i, e := range c.Errors {
				logger.Strict().Error(e.Err.Error(),defaultFields(c)...)
				errArray[i] = ErrMessage.GetErrorDescription(e.Err)
			}

			if c.IsAborted() {
				status = c.Writer.Status()
			} else if c := errArray[0].Code; c != 0 {
				status = c
			}

			c.JSON(status, gin.H{
				"errors": errArray,
			})
		}
	}
}

func defaultFields(c *gin.Context) (fs []zap.Field) {
	if tid := c.GetString(tracing.ContextHeaderName);tid != "" {
		fs = append(fs, zap.String(tracing.LoggerFieldKey,tid))
	}
	return fs
}
