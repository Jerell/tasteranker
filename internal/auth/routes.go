package auth

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
)

func UseSubroute(group *echo.Group) {
    group.GET("/:provider", func(c echo.Context) error {
        // Get provider from URL parameter
        provider := c.Param("provider")
        
        // Add provider to the request URL
        q := c.Request().URL.Query()
        q.Add("provider", provider)
        c.Request().URL.RawQuery = q.Encode()
        
        gothic.BeginAuthHandler(c.Response(), c.Request())
        return nil
    })
    
    // Update callback to also use the provider parameter
    group.GET("/:provider/callback", handleGoogleCallback)
    group.GET("/logout", handleLogout)
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

    // Add debug logging
    fmt.Printf("OAuth user data: ID=%s, Name=%s, Email=%s\n", user.UserID, user.Name, user.Email)

    session, _ := Store.Get(c.Request(), "auth-session")
    session.Values["user_id"] = user.UserID
    session.Values["email"] = user.Email
    session.Values["name"] = user.Name

    // Add debug logging for session
    fmt.Printf("Session values: %+v\n", session.Values)

    err = session.Save(c.Request(), c.Response().Writer)
    if err != nil {
        return err
    }

    return c.Redirect(http.StatusFound, "/")
}

func handleLogout(c echo.Context) error {
    session, err := Store.Get(c.Request(), "auth-session")
    if err != nil {
        return c.Redirect(http.StatusTemporaryRedirect, "/")
    }

    // Clear session values
    session.Values = map[interface{}]interface{}{}
    err = session.Save(c.Request(), c.Response().Writer)
    if err != nil {
        return err
    }

    return c.Redirect(http.StatusTemporaryRedirect, "/")
}
