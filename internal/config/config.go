package config

import (
	"log"
	"text/template"

	"github.com/alexedwards/scs/v2"
)

/* config is imported by other parts of the app
but it doesn't import  anything else from the app itself */

// AppConfig holds the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
}
