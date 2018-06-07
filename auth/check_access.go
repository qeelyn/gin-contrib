package auth

import "github.com/gin-gonic/gin"

type CheckAccess struct {
	CheckFunc         func(*gin.Context, string, string, map[string]interface{}) bool
	GetPermissionFunc func(*gin.Context) string
	UnauthorizedFunc  func(c *gin.Context, status int)
}

func (t *CheckAccess) CheckAccessHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		permission := t.GetPermissionFunc(c)
		if !t.CheckAccessExec(c, permission,map[string]interface{}{}) {
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

func (t *CheckAccess) CheckAccessExec(c *gin.Context, permission string,params map[string]interface{}) bool {
	userId := c.GetString("userId")
	if userId == "" {
		t.UnAuthorization(c, 401)
		return false
	}

	orgId := c.GetHeader("Qeelyn-Org-Id")
	if orgId != "" {
		params["org_id"] = orgId
	}
	if !t.CheckFunc(c, userId, permission, params) {
		t.UnAuthorization(c, 403)
		return false
	}
	return true
}
