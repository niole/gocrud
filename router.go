package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Text struct {
	Stuff string
}

func GetFormattedBody(req *http.Request) []FieldValue {
	decoder := json.NewDecoder(req.Body)
	values := make([]FieldValue, 0)

	for {
		var body map[string]interface{}
		err := decoder.Decode(&body)

		if err != nil {
			if err == io.EOF {
				return values
			} else {
				log.Fatal(err)
			}
		}

		for key, value := range body {

			foundFloat, ok := value.(float64)
			if ok {
				stringifiedNumber := strconv.FormatFloat(foundFloat, 'f', -1, 64)
				values = append(values, FieldValue{key, stringifiedNumber})
			} else {
				foundString, ok := value.(string)

				if ok {
					values = append(values, FieldValue{key, "'" + foundString + "'"})
				} else {
					log.Fatal("this is not a string. Must handle other type cases")
				}

			}

		}

	}

	defer req.Body.Close()

	return values
}

type Route struct {
	model  *Model
	cruder *Cruder
}

func (r *Route) Handler(w http.ResponseWriter, crudType string, values []FieldValue) {
	fmt.Println(w)
	fmt.Println(crudType)
	fmt.Println(values)

	//io.WriteString(w, "successful send")

	switch crudType {
	case "read":
		r.cruder.read(values)
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
		return
	}

	fieldValues := GetFormattedBody(r)
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
