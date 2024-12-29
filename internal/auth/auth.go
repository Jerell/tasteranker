package auth

import (
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func Init(baseURL string) {
    key := os.Getenv("SESSION_SECRET")
    if key == "" {
        panic("SESSION_SECRET environment variable is required")
    }
    
    maxAge := 86400 * 30 // 30 days
    store := sessions.NewCookieStore([]byte(key))
    store.MaxAge(maxAge)
    store.Options.Path = "/"
    store.Options.HttpOnly = true
    store.Options.Secure = os.Getenv("APP_ENV") != "development"
    
    gothic.Store = store
    
    goth.UseProviders(
        google.New(
            os.Getenv("GOOGLE_CLIENT_ID"),
            os.Getenv("GOOGLE_CLIENT_SECRET"),
            baseURL+"/auth/google/callback",
            "email", "profile", // Add required scopes
        ),
    )
}

func BeginAuthHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        gothic.BeginAuthHandler(w, r)
    }
}

func CallbackHandler(c echo.Context) error {
    user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    // Store user info in session
    session, err := gothic.Store.Get(c.Request(), gothic.SessionName)
    if err != nil {
        return err
    }

    session.Values["user_id"] = user.UserID
    session.Values["user_name"] = user.Name
    session.Values["user_email"] = user.Email
    
    err = session.Save(c.Request(), c.Response().Writer)
    if err != nil {
        return err
    }

    // Check if this is an HTMX request
    if c.Request().Header.Get("HX-Request") == "true" {
        // Return just the updated user info component
        return UserInfo(user.Name).Render(c.Request().Context(), c.Response().Writer)
    }

    // For regular requests, redirect to home page
    return c.Redirect(http.StatusFound, "/")
}

func LogoutHandler(c echo.Context) error {
    session, err := gothic.Store.Get(c.Request(), gothic.SessionName)
    if err != nil {
        return err
    }

    // Clear session
    session.Values = make(map[interface{}]interface{})
    err = session.Save(c.Request(), c.Response().Writer)
    if err != nil {
        return err
    }

    if c.Request().Header.Get("HX-Request") == "true" {
        // Return login button component for HTMX requests
        return LoginButton().Render(c.Request().Context(), c.Response().Writer)
    }

    return c.Redirect(http.StatusFound, "/")
}
