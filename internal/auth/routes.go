package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth/gothic"
)

func UseSubroute(group *echo.Group) {
    group.GET("/:provider", func(c echo.Context) error {
        provider := c.Param("provider")
        
        q := c.Request().URL.Query()
        q.Add("provider", provider)
        c.Request().URL.RawQuery = q.Encode()
        
        gothic.BeginAuthHandler(c.Response(), c.Request())
        return nil
    })
    
    group.GET("/:provider/callback", handleGoogleCallback)
    group.GET("/logout", handleLogout)
}

func handleGoogleAuth(c echo.Context) error {
    session, _ := gothic.Store.Get(c.Request(), gothic.SessionName)
    session.Values["provider"] = "google"
    err := session.Save(c.Request(), c.Response().Writer)
    if err != nil {
        return err
    }

    gothic.BeginAuthHandler(c.Response(), c.Request())
    return nil
}

func handleGoogleCallback(c echo.Context) error {
    user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
    if err != nil {
        return c.String(http.StatusInternalServerError, err.Error())
    }

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

func handleLogout(c echo.Context) error {
    session, err := Store.Get(c.Request(), "auth-session")
    if err != nil {
        return c.Redirect(http.StatusTemporaryRedirect, "/")
    }

    session.Values = map[interface{}]interface{}{}
    err = session.Save(c.Request(), c.Response().Writer)
    if err != nil {
        return err
    }

    return c.Redirect(http.StatusTemporaryRedirect, "/")
}
