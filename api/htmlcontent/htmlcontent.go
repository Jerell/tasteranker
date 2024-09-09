package htmlcontent

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func UseSubroute(group *echo.Group)  {
    group.GET("*", func (c echo.Context) error {
        key := c.Param("*")
		data := map[string]interface{}{
			"Name": key,
		}
        return c.Render(http.StatusOK, "hello.html", data)
    })
}
