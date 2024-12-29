package auth

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/markbates/goth/gothic"
)

func AuthContext(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        session, err := gothic.Store.Get(c.Request(), gothic.SessionName)
        if err != nil {
            // Set defaults if session error
            c.Set("authenticated", false)
            c.Set("user_name", "")
            return next(c)
        }

        _, ok := session.Values["user_id"]
        if !ok {
            c.Set("authenticated", false)
            c.Set("user_name", "")
            return next(c)
        }

        c.Set("authenticated", true)
        c.Set("user_name", session.Values["user_name"])
        return next(c)
    }
}

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        session, err := gothic.Store.Get(c.Request(), gothic.SessionName)
        if err != nil || session.Values["user_id"] == nil {
            return c.Redirect(http.StatusTemporaryRedirect, "/auth/google")
        }
        return next(c)
    }
}
