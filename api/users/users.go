package users

import (
	"net/http"

	"github.com/Jerell/tasteranker/components"
    "github.com/Jerell/tasteranker/internal/db"
    "github.com/Jerell/tasteranker/handlers"
	"github.com/labstack/echo/v4"
)

func UseSubroute(group *echo.Group, store *db.UserStore) {
    handler := handlers.NewUserHandler(store)

    group.GET("list", handler.List)
    group.POST("", handler.Create)
    
    group.GET("*", func(c echo.Context) error {
        key := c.Param("*")
        return components.Render(
            c, http.StatusOK,
            components.Main(
                components.Hello("user " + key),
            ),
        )
    })
}

