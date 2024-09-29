package home

import (
	"net/http"

	"github.com/Jerell/tasteranker/components"
	"github.com/labstack/echo/v4"
)

func UseSubroute(group *echo.Group) {
    group.GET("*", func(c echo.Context) error {
        key := c.Param("*")
        return components.Render(
            c, http.StatusOK,
            components.Main(
                components.Hello("udisjf " + key),
            ),
        )
    })
}
