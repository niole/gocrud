package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Text struct {
	Stuff string
}

func GetFormattedBody(req *http.Request) []FieldValue {
	decoder := json.NewDecoder(req.Body)

	var body map[string]string

	// while the array contains values
	for decoder.More() {
		err := decoder.Decode(&body)
		if err != nil {
			log.Fatal(err)
		}

	}
	defer req.Body.Close()

	values := make([]FieldValue, 0)
	for key, value := range body {
		values = append(values, FieldValue{key, value})
	}

	return values
}

type Route struct {
	model  *Model
	cruder *Cruder
}

func (r *Route) Handler(w http.ResponseWriter, crudType string, values []FieldValue) {
	fmt.Println(w)
	switch crudType {
	case "read":
		//r.cruder.read(values)
	case "update":
	case "create":
		r.cruder.create(values)
	case "delete":
	default:
		return
	}
}

type Router struct {
	routes        map[string]*Route
	pathValidator *regexp.Regexp
}

func (s *Router) DelegateRequest(w http.ResponseWriter, r *http.Request) {
	foundPath := s.pathValidator.FindStringSubmatch(r.URL.Path)

	if foundPath == nil {
		http.NotFound(w, r)
		log.Fatal("this route doesn't exist")
		return
	}

	fieldValues := GetFormattedBody(r)
	fmt.Println("fieldValues", fieldValues)
	crudType := foundPath[2]
	modelName := foundPath[1]
	route := s.routes[modelName]
	route.Handler(w, crudType, fieldValues)
}

func GenerateRouteValidator(models []*Model) *regexp.Regexp {
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
