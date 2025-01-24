package config

import (
	"log"
	"text/template"

	"github.com/alexedwards/scs/v2"
	"github.com/kons77/room-bookings-app/internal/models"
)

/* Config is imported by other parts of the app but it doesn't import anything else from the app itself.
Config is available to every part of the app  that has access to the app config. */

// AppConfig holds the application config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
	MailChan      chan models.MailData
}
