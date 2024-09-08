package main

import (
	"html/template"
	"io"
	"net/http"
    "embed"

	"github.com/Jerell/tasteranker/api/htmlcontent"
	"github.com/Jerell/tasteranker/components"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed public/views/*
var resources embed.FS

type Template struct {
    templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
    return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
    e := echo.New()

    t := &Template{
        templates: template.Must(template.ParseFS(resources, "public/views/*.html")),
    }
    e.Renderer = t

    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    e.GET("/t", func(c echo.Context) error {
        return components.Render(c, http.StatusOK, components.Hello("ooby"))

    })

    e.GET("/", func(c echo.Context) error {
        // Pass data to the template
        data := map[string]interface{}{
            "Name": "placeholder name",
        }
        return c.Render(http.StatusOK, "hello.html", data)
    })

    e.Static("/", "assets")

    htmlGroup := e.Group("/html")
    println(2020202002)
    htmlcontent.UseSubroute(htmlGroup)

    // Start server
    e.Logger.Fatal(e.Start(":80"))
}


