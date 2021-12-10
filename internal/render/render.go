package render

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/justinas/nosurf"
	"github.com/rdenman/bookings/internal/config"
	"github.com/rdenman/bookings/internal/models"
)

var functions = template.FuncMap{}

var app *config.AppConfig
func NewTemplates(a *config.AppConfig) {
	app = a
}

func AddDefaultData(td *models.TemplateData, req *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(req.Context(), "flash")
	td.Warning = app.Session.PopString(req.Context(), "warning")
	td.Error = app.Session.PopString(req.Context(), "error")
	td.CSRFToken = nosurf.Token(req)
	return td
}

func RenderTemplate(w http.ResponseWriter, req *http.Request, tmpl string, td *models.TemplateData) {
	var tc map[string]*template.Template
	if app.UseCache {
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		log.Fatal("could not get template from template cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, req)

	_ = t.Execute(buf, td)

	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writing template to browser", err)
	}
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		return cache, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return cache, err
		}

		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			return cache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return cache, err
			}
		}

		cache[name] = ts
	}

	return cache, nil
}