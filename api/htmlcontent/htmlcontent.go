package htmlcontent

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func UseSubroute(group *echo.Group)  {
    group.GET("", func (c echo.Context) error {
		data := map[string]interface{}{
			"Name": "Scooby",
		}
        return c.Render(http.StatusOK, "hello.html", data)
    })
}
