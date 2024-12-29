package components
import (
    "net/http"
    "github.com/a-h/templ"
    "github.com/labstack/echo/v4"
)

func Render(ctx echo.Context, status int, t templ.Component) error {
    err := t.Render(ctx.Request().Context(), ctx.Response().Writer)
    if err != nil {
        ctx.Logger().Error(err)
        return ctx.String(http.StatusInternalServerError, "failed to render response template")
    }
    return nil
}
