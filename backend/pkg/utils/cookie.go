package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthCookieName holds the login token (httpOnly, so scripts cannot read it).
const AuthCookieName = "access_token"

var cookieSecure = true

// SetCookieSecure sets whether the login cookie is HTTPS-only (COOKIE_SECURE;
// false only for local http://localhost runs).
func SetCookieSecure(secure bool) {
	cookieSecure = secure
}

// SetAuthCookie stores the token; maxAge in seconds, negative deletes it.
func SetAuthCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AuthCookieName, token, maxAge, "/", "", cookieSecure, true)
}
