package models

// store all the models that includes database models and the template data model

// TemplateData holds data sent from handlers to templates
type TemplateData struct {
	StringMap map[string]string
	IntMap    map[string]int
	FloatMap  map[string]float32
	Data      map[string]interface{}
	CSRFToken string // a cross-site request forgery token
	// message sending to users
	Flash   string
	Warning string
	Error   string
}
