package main

import (
	"html/template"
	"io"
	"net/http"
	"runtime"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type TemplateRenderer struct {
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}
	templateContents, err := rice.MustFindBox("templates").String(name)
	if err != nil {
		return err
	}
	tmpl, err := template.New(name).Parse(templateContents)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

func getIndex(c echo.Context) error {
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{"title": "Yo Bot!"})
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:      "1; mode=block",
		ContentTypeNosniff: "nosniff",
		XFrameOptions:      "SAMEORIGIN",
		HSTSMaxAge:         63072000,
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"method":"${method}","uri":"${uri}","status":${status},` +
			`"latency_human":"${latency_human}","bytes_in":${bytes_in},` +
			`"bytes_out":${bytes_out}}` + "\n",
	}))

	e.HideBanner = true
	e.Renderer = &TemplateRenderer{}

	staticFiles := http.FileServer(rice.MustFindBox("static").HTTPBox())
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", staticFiles)))
	e.GET("/", getIndex)

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 200

	s := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           e.Server.Handler,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	s.SetKeepAlivesEnabled(false)
	e.Logger.Fatal(s.ListenAndServe())
}
