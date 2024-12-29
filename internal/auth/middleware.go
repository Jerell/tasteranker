package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func AuthContext(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        session, err := Store.Get(c.Request(), "auth-session")
        if err != nil {
            fmt.Printf("Session error: %v\n", err)
            c.Set("authenticated", false)
            c.Set("user_name", "")
            return next(c)
        }

        fmt.Printf("Session values in middleware: %+v\n", session.Values)
        
        _, ok := session.Values["user_id"]
        if !ok {
            fmt.Printf("No user_id found in session\n")
            // Try with context.WithValue
            ctx := context.WithValue(c.Request().Context(), "authenticated", false)
            ctx = context.WithValue(ctx, "user_name", "")
            c.SetRequest(c.Request().WithContext(ctx))
            // Also keep the c.Set for debugging
            c.Set("authenticated", false)
            c.Set("user_name", "")
            return next(c)
        }

        // Try both methods of setting the context
        fmt.Printf("Setting context values for authenticated user: %v\n", session.Values["name"])
        ctx := context.WithValue(c.Request().Context(), "authenticated", true)
        ctx = context.WithValue(ctx, "user_name", session.Values["name"])
        c.SetRequest(c.Request().WithContext(ctx))
        // Also keep the c.Set for debugging
        c.Set("authenticated", true)
        c.Set("user_name", session.Values["name"])

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
