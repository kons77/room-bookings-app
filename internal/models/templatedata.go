package models

import "github.com/kons77/room-bookings-app/internal/forms"

// store all the models that includes database models and the template data model

// TemplateData holds data sent from handlers to templates
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32
	Data      map[string]interface{}
	CSRFToken string // a cross-site request forgery token
	Flash     string // message sending to users
	Warning   string // message sending to users
	Error     string // message sending to users
	Form      *forms.Form
}
