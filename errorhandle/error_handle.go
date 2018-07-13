package errorhandle

import (
	"github.com/gin-gonic/gin"
	"github.com/qeelyn/go-common/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/qeelyn/go-common/logger"
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
				logger.GetZap().Error(e.Err.Error())
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
