package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func UseSubroute(group *echo.Group) {
    // Add OPTIONS handler
    group.OPTIONS("/google", func(c echo.Context) error {
        return c.NoContent(http.StatusOK)
    })
    
    group.GET("/google", func(c echo.Context) error {
        // Set provider in query string
        q := c.Request().URL.Query()
        q.Set("provider", "google")
        c.Request().URL.RawQuery = q.Encode()
        
        // Set required headers for CORS
        c.Response().Header().Set("Access-Control-Allow-Origin", c.Request().Header.Get("Origin"))
        c.Response().Header().Set("Access-Control-Allow-Credentials", "true")
        
        handler := BeginAuthHandler()
        handler.ServeHTTP(c.Response(), c.Request())
        return nil
    })
    
    // Add OPTIONS handler for callback
    group.OPTIONS("/google/callback", func(c echo.Context) error {
        return c.NoContent(http.StatusOK)
    })
    
    group.GET("/google/callback", CallbackHandler)
    group.POST("/logout", LogoutHandler)
}
