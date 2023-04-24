package main

import (
	"fmt"
	"io"
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var e = createMux()

func createMux() *echo.Echo {
	e := echo.New()
	return e
}

type TemplateRender struct {
	templates *template.Template
}

func (t *TemplateRender) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	cfg, err := NewConfig()
	if err != nil {
		println(err)
		return
	}
	fmt.Printf("%#v", cfg)

	// SSL証明書検証を無効化
	// http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Echoインスタンス作成 //
	http.Handle("/", e)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Gzip())

	// Template
	renderer := &TemplateRender{
		templates: template.Must(template.ParseGlob("view/*.html")),
	}
	e.Renderer = renderer

	// ルーティング //
	handler := NewHandler(cfg)
	e.GET("/", handler.IndexHandler)
	e.GET("/ip", handler.IpHandler)
	e.Start(":" + cfg.Port)
}
