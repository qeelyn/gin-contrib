package auth

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/qeelyn/go-common/auth"
	"net/http"
	"strings"
	"time"
)

var(
	// ErrEmptyAuthHeader can be thrown if authing with a HTTP header, the Auth header needs to be set
	ErrEmptyAuthHeader = errors.New("auth header is empty")

	// ErrInvalidAuthHeader indicates auth header is invalid, could for example have the wrong Realm name
	ErrInvalidAuthHeader = errors.New("auth header is invalid")

	// ErrEmptyQueryToken can be thrown if authing with URL Query, the query token variable is empty
	ErrEmptyQueryToken = errors.New("query token is empty")

	// ErrEmptyCookieToken can be thrown if authing with a cookie, the token cokie is empty
	ErrEmptyCookieToken = errors.New("cookie token is empty")
)

// GinJWTMiddleware provides a Json-Web-Token authentication implementation. On failure, a 401 HTTP response
// is returned. On success, the wrapped middleware is called, and the userID is made available as
// c.Get("userid").(string).
// Users can get a token by posting a json request to LoginHandler. The token then needs to be passed in
// the Authentication header. Example: Authorization:Bearer XXX_TOKEN_XXX
type GinJWTMiddleware struct {
	*auth.BearerTokenValidator
	// signing algorithm - possible values are HS256, HS384, HS512
	// Optional, default is HS256.
	SigningAlgorithm string

	// Duration that a jwt token is valid. Optional, defaults to one hour.
	Timeout time.Duration
	// Callback function that should perform the authorization of the authenticated user. Called
	// only after an authentication success. Must return true on success, false on failure.
	// Optional, default to success.
	TokenValidator func(token *jwt.Token, c *gin.Context) bool

	// User can define own UnauthorizedFunc func.if return true,request will abort
	UnauthorizedHandle func(*gin.Context, int, string) bool

	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup string

	// TokenHeadName is a string in the header. Default value is "Bearer"
	TokenHeadName string

	// HTTP Status messages for when something in the JWT middleware fails.
	// Check error (e) to determine the appropriate error message.
	HTTPStatusMessageFunc func(e error, c *gin.Context) string
}

func (t *GinJWTMiddleware) usingPublicKeyAlgo() bool {
	switch t.SigningAlgorithm {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

// Init initialize jwt configs.
func (t *GinJWTMiddleware) Init() error {

	if t.TokenLookup == "" {
		t.TokenLookup = "header:Authorization"
	}

	if t.SigningAlgorithm == "" {
		t.SigningAlgorithm = "HS256"
	}

	t.TokenHeadName = strings.TrimSpace(t.TokenHeadName)
	if len(t.TokenHeadName) == 0 {
		t.TokenHeadName = "Bearer"
	}

	if t.TokenValidator == nil {
		t.TokenValidator = func(token *jwt.Token, c *gin.Context) bool {
			return true
		}
	}

	if t.UnauthorizedHandle == nil {
		t.UnauthorizedHandle = func(c *gin.Context, code int, message string) bool {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
			return true
		}
	}

	if t.IdentityHandler == nil {
		t.IdentityHandler = func(c context.Context, claims jwt.MapClaims) (*auth.Identity, error) {
			id := claims["sub"].(string)
			return &auth.Identity{
				Id: id,
			}, nil
		}
	}

	if t.HTTPStatusMessageFunc == nil {
		t.HTTPStatusMessageFunc = func(e error, c *gin.Context) string {
			return e.Error()
		}
	}

	if t.Realm == "" {
		return auth.ErrMissingRealm
	}

	if err := t.BearerTokenValidator.Init(); err != nil {
		return err
	}

	return nil
}

// Handle makes GinJWTMiddleware implement the Middleware interface.
func (t *GinJWTMiddleware) Handle() gin.HandlerFunc {
	if err := t.Init(); err != nil {
		return func(c *gin.Context) {
			t.unauthorized(c, http.StatusInternalServerError, t.HTTPStatusMessageFunc(err, nil))
			return
		}
	}

	return func(c *gin.Context) {
		t.middlewareImpl(c)
		return
	}
}

func (t *GinJWTMiddleware) middlewareImpl(c *gin.Context) {
	token, err := t.getTokenString(c)

	if err != nil {
		t.unauthorized(c, http.StatusUnauthorized, t.HTTPStatusMessageFunc(err, c))
		return
	}

	identity,err := t.Validate(c,token)
	if err != nil {
		t.unauthorized(c, http.StatusUnauthorized, t.HTTPStatusMessageFunc(err, c))
		return
	}

	c.Set("userid", identity.Id)

	c.Next()
}

func (t *GinJWTMiddleware) signedString(token *jwt.Token) (string, error) {
	var tokenString string
	var err error
	if t.usingPublicKeyAlgo() {
		tokenString, err = token.SignedString(t.PrivKeyFile)
	} else {
		tokenString, err = token.SignedString(t.Key)
	}
	return tokenString, err
}

// TokenGenerator method that clients can use to get a jwt token.
func (t *GinJWTMiddleware) TokenGenerator(userID string) (string, time.Time, error) {

	token := jwt.New(jwt.GetSigningMethod(t.SigningAlgorithm))
	expire := time.Now().Add(t.Timeout)
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": expire.Unix(),
	}
	token.Claims = claims
	tokenString, err := t.signedString(token)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expire, nil
}

func (t *GinJWTMiddleware) jwtFromHeader(c *gin.Context, key string) (string, error) {
	authHeader := c.Request.Header.Get(key)

	if authHeader == "" {
		return "", ErrEmptyAuthHeader
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == t.TokenHeadName) {
		return "", ErrInvalidAuthHeader
	}

	return parts[1], nil
}

func (t *GinJWTMiddleware) jwtFromQuery(c *gin.Context, key string) (string, error) {
	token := c.Query(key)

	if token == "" {
		return "", ErrEmptyQueryToken
	}

	return token, nil
}

func (t *GinJWTMiddleware) jwtFromCookie(c *gin.Context, key string) (string, error) {
	cookie, _ := c.Cookie(key)

	if cookie == "" {
		return "", ErrEmptyCookieToken
	}

	return cookie, nil
}

func (t *GinJWTMiddleware) getTokenString(c *gin.Context) (token string, err error) {
	parts := strings.Split(t.TokenLookup, ":")
	switch parts[0] {
	case "header":
		token, err = t.jwtFromHeader(c, parts[1])
	case "query":
		token, err = t.jwtFromQuery(c, parts[1])
	case "cookie":
		token, err = t.jwtFromCookie(c, parts[1])
	}
	return
}

func (t *GinJWTMiddleware) unauthorized(c *gin.Context, code int, message string) {
	if t.UnauthorizedHandle(c, code, message) {
		if t.Realm == "" {
			t.Realm = "gin jwt"
		}

		c.Header("WWW-Authenticate", "JWT realm="+t.Realm)
		c.Abort()
	}
}
