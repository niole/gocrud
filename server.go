package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)

	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	} else {
		renderTemplate(w, p, "view.html")
	}
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}

	renderTemplate(w, p, "edit.html")
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	p.save()
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func renderTemplate(w http.ResponseWriter, p *Page, tmpl string) {
	err := templates.ExecuteTemplate(w, tmpl, p)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalRouterError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)

		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

type Text struct {
	Stuff string
}

func GetFormattedBody(req *http.Request) []FieldValue {
	decoder := json.NewDecoder(req.Body)

	var body interface{}

	err := decoder.Decode(&body)
	if err != nil {
		log.Fatal(err)
	}
	defer req.Body.Close()

	values := make([]FieldValue, 0)
	for key, value := range body {
		values = append(values, FieldValue{key, string(value)})
	}

	return values
}

type Route struct {
	model  *Model
	cruder Cruder
}

func (r *Route) Handler(crudType string) {
	switch crudType {
	case "read":
		r.cruder.read()
	case "update":
	case "create":
	case "delete":
	default:
		return
	}
}

type Router struct {
	routes        map[string]*Route
	pathValidator *Regexp
}

func (s *Router) DelegateRequest(w http.ResponseWriter, r *http.Request) {
	foundPath := s.pathValidator.FindStringSubmatch(r.URL.Path)

	if foundPath == nil {
		http.NotFound(w, r)
		log.Fatal("this route doesn't exist")
		return
	}

	crudType := foundPath[2]
	modelName := foundPath[1]
	route := s.routes[modelName]
	route.Handler(crudType)
}

func GenerateRouteValidator(models []*Model) *Regexp {
	baseChecker := make([]string, len(models))
	crudTypeChecker := "(create|read|update|delete)"

	for i, model := range models {
		baseChecker[i] = model.GetName()
	}

	regexpContent := "^/(" + strings.Join(baseChecker, "|") + ")/" + crudTypeChecker + "$"
	return regexp.MustCompile(regexpContent)
}

func InitRouter(db *DataBase, models []*Model) *Router {
	routes := make(map[string]*Route, 0)
	cruders := InitCruders(db, models)

	for _, model := range models {
		modelName := model.GetName()
		routes[modelName] = &Route{model, cruders[modelName]}
	}

	return &Router{routes, GenerateRouteValidator(models)}
}
