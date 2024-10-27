package main

import (
	"context"
	"embed"
	"html/template"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Jerell/tasteranker/api/htmlcontent"
	"github.com/Jerell/tasteranker/api/users"
	"github.com/Jerell/tasteranker/components"
	"github.com/Jerell/tasteranker/tigris"
    "github.com/Jerell/tasteranker/internal/db"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/joho/godotenv"
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

    err := godotenv.Load()
    if err != nil {
        e.Logger.Fatal("Error loading .env file")
    }

    dbConfig := db.NewConfig()
    database, err := db.NewConnection(dbConfig)
    if err != nil {
        e.Logger.Fatal(err)
    }
    defer database.Close()

    env := os.Getenv("APP_ENV")

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    t := &Template{
        templates: template.Must(
            template.ParseFS(
                resources,
                "public/views/*.html",
            ),
        ),
    }
    e.Renderer = t

    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    if env == "development" {
        e.Static("/assets", "./assets")
    } else {
        e.GET("/assets/*", func(c echo.Context) error {
            key := c.Param("*")
            ctx := context.Background()

            client, err := tigris.Client( ctx )
            if err != nil {
                return c.String(http.StatusInternalServerError, "Failed to create Tigris client")
            }

            // Use the Tigris client to get the object
            resp, err := client.GetObject(ctx, &s3.GetObjectInput{
                Bucket: aws.String("frosty-sound-5710"),
                Key:    aws.String("assets/"+key),
            })
            if err != nil {
                return c.String(http.StatusNotFound, "File not found")
            }
            defer resp.Body.Close()

            contentType := mime.TypeByExtension(filepath.Ext(key))
            if contentType == "" {
                contentType = "application/octet-stream" // Default fallback
            }

            // Stream the S3 object to the client
            return c.Stream(http.StatusOK, contentType, resp.Body)
        })
    }

    e.GET("/", func(c echo.Context) error {
        return components.Render(
            c, http.StatusOK, 
            components.Main(components.Home()),
        )
    })

    usersGroup := e.Group("/users/")
    userStore := db.NewUserStore(database)
    users.UseSubroute(usersGroup, userStore)

    htmlGroup := e.Group("/html/")
    htmlcontent.UseSubroute(htmlGroup)

    e.Logger.Fatal(e.Start(":"+port))
}


