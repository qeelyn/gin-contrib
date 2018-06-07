package errorhandle

import (
	"github.com/gin-gonic/gin"
	"github.com/qeelyn/go-common/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/qeelyn/gin-contrib/ginzap"
)

func ErrorHandle(config map[string]interface{},logger *ginzap.Logger) gin.HandlerFunc {
	bytes, err := ioutil.ReadFile(config["error-template"].(string))
	if err != nil {
		panic(err)
	}
	templates := map[string]errors.ErrorTemplate{}
	yaml.Unmarshal(bytes, &templates)
	eh := errors.NewErrorHandle(&templates)

	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			var (
				errArray = []*errors.ErrorDescription{}
				status   = 500
			)
			if c.IsAborted() {
				status = c.Writer.Status()
			}

			for _, e := range c.Errors {
				logger.GetZap().Error(e.Err.Error())
				errArray = append(errArray, eh.GetErrorDescription(e.Err))
				if errArray[0].Code != 0 {
					status = errArray[0].Code
				}
			}

			c.JSON(status, gin.H{
				"errors": errArray,
			})
		}
	}
}
