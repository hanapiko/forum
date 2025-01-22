package handlers

import (
	"html/template"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// TemplateData is a generic struct for passing data to templates
type TemplateData struct {
	PageTitle   string
	User        interface{}
	CurrentYear int
	Error       string
	SuccessMsg  string
	Data        interface{}
}

// TemplateRenderer handles parsing and rendering of HTML templates
type TemplateRenderer struct {
	templates      *template.Template
	authMiddleware AuthMiddleware
}

// NewTemplateRenderer creates a new template renderer
func NewTemplateRenderer(authMiddleware AuthMiddleware) *TemplateRenderer {
	// Parse all templates, including base and component templates
	templates := template.Must(template.New("").Funcs(template.FuncMap{
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		"formatDate": func(t time.Time) string {
			return t.Format("Jan 2, 2006")
		},
	}).ParseGlob("frontend/templates/**/*.html"))

	return &TemplateRenderer{templates: templates, authMiddleware: authMiddleware}
}

// Render renders a specific template with given data
func (tr *TemplateRenderer) Render(w http.ResponseWriter, r *http.Request, templateName string, data TemplateData) error {
	// Set default values
	if data.CurrentYear == 0 {
		data.CurrentYear = time.Now().Year()
	}

	// Add user from context if not set
	if data.User == nil {
		data.User = r.Context().Value("user")
	}

	// Execute the base template with the specific content template
	err := tr.templates.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}

// RenderError renders an error template
func (tr *TemplateRenderer) RenderError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	data := TemplateData{
		PageTitle: "Error",
		Error:     err.Error(),
	}

	w.WriteHeader(statusCode)

	switch statusCode {
	case http.StatusForbidden:
		tr.Render(w, r, "errors/403", data)
	case http.StatusInternalServerError:
		tr.Render(w, r, "errors/500", data)
	default:
		tr.Render(w, r, "errors/500", data)
	}
}

// ParseTemplates is a utility function to manually parse templates if needed
func (tr *TemplateRenderer) ParseTemplates() error {
	templates, err := template.ParseGlob("frontend/templates/**/*.html")
	if err != nil {
		return err
	}
	tr.templates = templates
	return nil
}

// RenderTemplate renders a template with dynamic data and authentication context
func (tr *TemplateRenderer) RenderTemplate(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	// Create a copy of the data to avoid modifying the original
	renderData := make(map[string]interface{})

	// Convert input data to map if it's not already
	if v, ok := data.(map[string]interface{}); ok {
		renderData = v
	} else {
		val := reflect.ValueOf(data)
		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			renderData[typ.Field(i).Name] = val.Field(i).Interface()
		}
	}

	// Add common template data
	renderData["Title"] = strings.Title(strings.ReplaceAll(templateName, "/", " "))
	renderData["Year"] = time.Now().Year()
	renderData["IsAuthenticated"] = tr.authMiddleware.IsAuthenticated(r)

	// Render template
	tmpl, err := tr.parseTemplate(templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, renderData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// parseTemplate parses a template by name
func (tr *TemplateRenderer) parseTemplate(templateName string) (*template.Template, error) {
	tmpl := tr.templates.Lookup(templateName)
	if tmpl == nil {
		return nil, fmt.Errorf("template %s not found", templateName)
	}
	return tmpl, nil
}

// AuthMiddleware is an interface for authentication middleware
type AuthMiddleware interface {
	IsAuthenticated(r *http.Request) bool
}
