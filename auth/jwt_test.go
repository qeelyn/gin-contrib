package auth_test

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/qeelyn/gin-contrib/auth"
	auth2 "github.com/qeelyn/go-common/auth"
	"net/http"
	"testing"
	"time"
)

var(
	algorithm = "RS256"
	priKey = []byte(
		`-----BEGIN PRIVATE KEY-----
MIIBVgIBADANBgkqhkiG9w0BAQEFAASCAUAwggE8AgEAAkEA4b00JtFMZTavmKq/
3MdC5XkFh3NfAj7uzJS47WC3o42pY3T/hvaVbQF2OGuypcu85w+akSvSUKw94xwJ
qgeecwIDAQABAkAfgMIYcLkSnvEegyXHb998GsmUv5sQuyukTVUJe6flIQrZd69m
Q7VY4CYqDxw2jLtv1rqekfB9pdk7k048D9TBAiEA9FTcyia/g/C0ikYJLoCtNjDr
jAtYclYTeHmZJg/YE/ECIQDshQjt8iA9MK9AMSjZMHT1/DlfYOsxahnWjvKFH92s
owIhAMP4UQLfI1sbNGN3myOuV7+Qa0zvSKikO4e02E58BM6xAiEAzNVF73RSkUu5
aploa/f4QxRlx4FTDp95swRnaf036IsCIQCHPiMr1lEIDz/cDL1+sA5CUbep5TpC
jhCXtOqX4BVBWg==
-----END PRIVATE KEY-----
`	)
	pubKey = []byte(
		`-----BEGIN PUBLIC KEY-----
MFwwDQYJKoZIhvcNAQEBBQADSwAwSAJBAOG9NCbRTGU2r5iqv9zHQuV5BYdzXwI+
7syUuO1gt6ONqWN0/4b2lW0BdjhrsqXLvOcPmpEr0lCsPeMcCaoHnnMCAwEAAQ==
-----END PUBLIC KEY-----
`	)
	ekey = "passw0rd"
)

func TestGinJWTMiddleware_Handle(t *testing.T) {
	// the jwt middleware
	am := &auth.GinJWTMiddleware{
		BearerTokenValidator: &auth2.BearerTokenValidator{
			Realm:            "test",
			PrivKeyFile:      priKey,
			PubKeyFile:       pubKey,
			Key:              []byte(ekey),
		},
		SigningAlgorithm: algorithm,
		Timeout:          1 * time.Second,
		UnauthorizedHandle: func(c *gin.Context, code int, message string) bool {
			c.Error(errors.New(message))
			// gin http response will cause crash
			return false
		},
		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TokenValidator: func(token *jwt.Token, c *gin.Context) bool {
			claims := token.Claims.(jwt.MapClaims)
			if claims["jti"] == nil {
				// the jwt is from login handle
				return true
			} else {
				return false
			}
		},
	}
	// no expire token
	aheader := `Bearer eyJhbGciOiJSUzI1NiJ9.eyJpYXQiOjE1MzgwNTE4ODAsInN1YiI6IjEyMiJ9.bgu0JhkL8ocPGExoATtyx6qRUxjT_ghz8EYaPh_sqBfqliy7mAwkg7OUjTUlv8fxwqy_1WvtyS8ZmYoFfv6ABQ`
	c := &gin.Context{
		Request: &http.Request{
			Header: map[string][]string{
				"Authorization": {aheader},
			},
		},
	}
	am.Handle()(c)
	if c.Errors != nil {
		t.Fatal(c.Errors.Errors())
	}
	// expire token
	bheader := `Bearer eyJhbGciOiJSUzI1NiJ9.eyJleHAiOjE1MzgwNTI4ODAsImlhdCI6MTUzODA1MTg4MCwic3ViIjoiMTIyIn0.WRF2bYUb-R3JzJJ0NuqrgpaLGzCRmDjLdCpJQGZ7-TPGIDpb01JvEdAyWlE060zI2BZ5s9EWkc6_G4hk_JDPhA`
	c1 := &gin.Context{
		Request: &http.Request{
			Header: map[string][]string{
				"Authorization": {bheader},
			},
		},
	}
	am.Handle()(c1)
	if c1.Errors == nil {
		t.Fatal("it should be happen error")
	}
	//bad
	cheader := `Bearer eyJhbGciOiJSUzI1NiJ9.eyJleHAiOjE1MzgwNTI4ODAsImlhdCI6MTUzODA1MTg4MCwic3ViIjoiMTIyIn0.WRF2bYUb-R3JzJJ0NuqrgpaLGzCRmDjLdCpJQGZ7-TPGIDpb01JvEdAyWlE060zI2BZ5s9EWkc6_G4hk_JDPhA-Bad`
	c2 := &gin.Context{
		Request: &http.Request{
			Header: map[string][]string{
				"Authorization": {cheader},
			},
		},
	}
	am.Handle()(c2)
	if c2.Errors == nil {
		t.Fatal("it should be happen error")
	}

}
