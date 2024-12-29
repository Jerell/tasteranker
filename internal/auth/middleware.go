package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// SetContextValues sets values in both the request context and Echo context
func SetContextValues(c echo.Context, values map[string]interface{}) {
	ctx := c.Request().Context()

	for key, value := range values {
		ctx = context.WithValue(ctx, key, value)
		c.Set(key, value) // also set in Echo's context
	}

	c.SetRequest(c.Request().WithContext(ctx))
}

func AuthContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := Store.Get(c.Request(), "auth-session")
		if err != nil {
			SetContextValues(c, map[string]interface{}{
				"authenticated": false,
				"user_name":     "",
				"user_id":       "",
			})
			return next(c)
		}

		fmt.Printf("Session values in middleware: %+v\n", session.Values)

		userID, ok := session.Values["user_id"]
		if !ok {
			SetContextValues(c, map[string]interface{}{
				"authenticated": false,
				"user_name":     "",
				"user_id":       "",
			})
            return next(c)
		}

        SetContextValues(c, map[string]interface{}{
            "authenticated": true,
            "user_name":    session.Values["name"],
            "user_id": userID,
        })

		return next(c)
	}
}

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := Store.Get(c.Request(), "auth-session")
		if err != nil || session.Values["user_id"] == nil {
			return c.Redirect(http.StatusTemporaryRedirect, "/auth/google")
		}
		return next(c)
	}
}
