package auth

import "github.com/gin-gonic/gin"

type CheckAccess struct {
	CheckFunc         func(*gin.Context, string, string, map[string]interface{}) bool
	GetPermissionFunc func(*gin.Context) string
	UnauthorizedFunc  func(c *gin.Context, status int)
}

func (t *CheckAccess) CheckAccessHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.GetString("userId")
		if userId == "" {
			t.UnAuthorization(c, 401)
			return
		}
		permission := t.GetPermissionFunc(c)
		orgId := c.GetHeader("Qeelyn-Org-Id")
		params := map[string]interface{}{}
		if orgId != "" {
			params["org_id"] = orgId
		}
		if !t.CheckFunc(c, userId, permission, params) {
			t.UnAuthorization(c, 403)
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
