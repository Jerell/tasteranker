package auth

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/markbates/goth/gothic"
)

func UseSubroute(group *echo.Group) {
    group.GET("/google", handleGoogleAuth)
    group.GET("/google/callback", handleGoogleCallback)
}

func handleGoogleAuth(c echo.Context) error {
    // Set provider in session
    session, _ := gothic.Store.Get(c.Request(), gothic.SessionName)
    session.Values["provider"] = "google"
    err := session.Save(c.Request(), c.Response().Writer)
    if err != nil {
        return err
    }

    // Call BeginAuthHandler and return nil since it handles the response internally
    gothic.BeginAuthHandler(c.Response(), c.Request())
    return nil  // Add explicit return
}

func handleGoogleCallback(c echo.Context) error {
    user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

    // Create and save user session
    session, _ := Store.Get(c.Request(), "auth-session")
    session.Values["user_id"] = user.UserID
    session.Values["email"] = user.Email
    session.Values["name"] = user.Name
    err = session.Save(c.Request(), c.Response().Writer)
    if err != nil {
        return err
    }

    return c.Redirect(http.StatusFound, "/")
}
