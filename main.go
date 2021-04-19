package main

import (
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
)

type CachedTemplateRenderer struct {
	m         sync.Mutex
	templates map[string]*template.Template
}

func NewTemplateRenderer() *CachedTemplateRenderer {
	return &CachedTemplateRenderer{
		templates: map[string]*template.Template{},
	}
}

func (t *CachedTemplateRenderer) getTemplate(name string) (*template.Template, error) {
	t.m.Lock()
	defer t.m.Unlock()
	tmpl, ok := t.templates[name]
	if !ok {
		templateContents, err := rice.MustFindBox("templates").String(name)
		if err != nil {
			return nil, err
		}
		tmpl, err = template.New(name).Parse(templateContents)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}
		t.templates[name] = tmpl
	}
	return tmpl, nil
}

func (t *CachedTemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if viewContext, isMap := data.(map[string]interface{}); isMap {
		viewContext["reverse"] = c.Echo().Reverse
	}
	tmpl, err := t.getTemplate(name)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

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

type ChoiceTemplateRenderer struct {
	cached *CachedTemplateRenderer
	normal *TemplateRenderer
}

func (t *ChoiceTemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if c.QueryParam("cached") != "" {
		return t.cached.Render(w, name, data, c)
	}
	return t.normal.Render(w, name, data, c)
}

type data struct {
	Name string
	Age  int
	Job  string
}

func randomString(num int) string {
	return random.String(uint8(num), random.Alphabetic)
}

func randomInt() int {
	return rand.Intn(100)
}

func getIndex(c echo.Context) error {
	var size = 200
	if s := c.QueryParam("size"); s != "" {
		size, _ = strconv.Atoi(s)
	}
	rows := make([]data, 0, size)
	for i := 0; i < size; i++ {
		rows = append(rows, data{Name: randomString(20), Age: randomInt(), Job: randomString(40)})
	}
	return c.Render(http.StatusOK, "index.html", map[string]interface{}{"title": "Yo Bot!", "rows": rows})
}

func gc(c echo.Context) error {
	debug.FreeOSMemory()
	return c.HTML(200, "done")
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
	e.Renderer = &ChoiceTemplateRenderer{
		cached: NewTemplateRenderer(),
		normal: &TemplateRenderer{},
	}

	staticFiles := http.FileServer(rice.MustFindBox("static").HTTPBox())
	e.GET("/static/*", echo.WrapHandler(http.StripPrefix("/static/", staticFiles)))
	e.GET("/", getIndex)
	e.GET("/gc", gc)

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 200

	s := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           e.Server.Handler,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      100 * time.Second,
		IdleTimeout:       10 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	s.SetKeepAlivesEnabled(false)
	e.Logger.Fatal(s.ListenAndServe())
}
