package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type CheckAccess struct {
	CheckFunc         func(*http.Request, string, string, map[string]interface{}) int
	GetPermissionFunc func(*gin.Context) string
	UnauthorizedFunc  func(c *gin.Context, status int)
}

func (t *CheckAccess) CheckAccessHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		permission := t.GetPermissionFunc(c)
		if !t.CheckAccessExec(c, permission, map[string]interface{}{}) {
			return
		}
		c.Next()
	}
}

func (t *CheckAccess) UnAuthorization(c *gin.Context, status int) {
	if t.UnauthorizedFunc != nil {
		t.UnauthorizedFunc(c, status)
	} else {
		c.AbortWithStatus(status)
	}
}

func (t *CheckAccess) CheckAccessExec(c *gin.Context, permission string, params map[string]interface{}) bool {
	userId := c.GetString("userId")
	if userId == "" {
		t.UnAuthorization(c, 401)
		return false
	}

	orgId := c.GetString("orgid")
	if orgId != "" {
		params["org_id"] = orgId
	}
	if sc := t.CheckFunc(c.Request, userId, permission, params); sc != http.StatusOK {
		t.UnAuthorization(c, sc)
		return false
	}
	return true
}
