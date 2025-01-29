package render

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"text/template"
	"time"

	"github.com/justinas/nosurf"
	"github.com/kons77/room-bookings-app/internal/config"
	"github.com/kons77/room-bookings-app/internal/models"
)

// holds all of the functions that we want to put into or make available to Goland templates.
var functions = template.FuncMap{
	"humanDate":  HumanDate,
	"formatDate": FormatDate,
	"iterate":    Iterate,
	// "add": Add,
}

var app *config.AppConfig
var pathToTemplates = "./templates"

/*
func Add(a,b int) int{
	return a+b
}
*/

// Iterate returns a slice of ints starting at 1, going to count
func Iterate(count int) []int {
	var i int
	var items []int
	for i = 1; i < count; i++ {
		items = append(items, i)
	}
	return items
}

// NewRenderer sets the config for the template package
func NewRenderer(a *config.AppConfig) {
	app = a
}

// HumanDate returns time in YYYY-MM-DD format
func HumanDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDate shows date as a month or as a year
func FormatDate(t time.Time, f string) string {
	return t.Format(f)
}

// AddDefaultData adds data for all templates (displayed on every page)
func AddDefaultData(td *models.TemplateData, r *http.Request) *models.TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	td.CSRFToken = nosurf.Token(r)
	if app.Session.Exists(r.Context(), "user_id") {
		td.IsAuthenticated = 1
	}
	return td
}

// Template renders a template - old
func Template(w http.ResponseWriter, r *http.Request, tmpl string, td *models.TemplateData) error {

	var tc map[string]*template.Template
	if app.UseCache {
		// get the template cache from the app config
		tc = app.TemplateCache
	} else {
		tc, _ = CreateTemplateCache()
	}

	// get requested template from chache
	t, ok := tc[tmpl]
	if !ok {
		//log.Fatal("can't get template from template cache")
		return errors.New("can't get template from template cache")
	}

	buf := new(bytes.Buffer)

	td = AddDefaultData(td, r)

	// the value I got from the map
	_ = t.Execute(buf, td)

	// render the template
	_, err := buf.WriteTo(w)
	if err != nil {
		fmt.Println("Error writeng template to browser", err)
		return err
	}

	return nil
}

// It's no longer need to keep track of what files are inside templates folder but still read files form FS every run
// CreateTemplateCache creates a template cache as a map
func CreateTemplateCache() (map[string]*template.Template, error) {
	// myCache := make(map[string]*template.Template)
	myCache := map[string]*template.Template{}

	// get all of the files *.page.tmpl from./templates
	pages, err := filepath.Glob(fmt.Sprintf("%s/*.page.tmpl", pathToTemplates))
	if err != nil {
		return myCache, err
	}

	// range through all files ending with *page.tmpl
	for _, page := range pages {
		// name of the file minus full path to it
		name := filepath.Base(page)
		// Funcs(function) позволяет добавлять пользовательские функции, которые можно использовать внутри шаблона
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return myCache, err
		}

		matches, err := filepath.Glob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
		if err != nil {
			return myCache, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob(fmt.Sprintf("%s/*.layout.tmpl", pathToTemplates))
			if err != nil {
				return myCache, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
