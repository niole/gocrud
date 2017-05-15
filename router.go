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

const (
	WHERE_CLAUSE = "where"
)

type CrudRequest struct {
	where []FieldFilter
	main  []FieldValue
}

func (c *CrudRequest) GetValues() []FieldValue {
	return c.main
}

func (c *CrudRequest) GetFilters() []FieldFilter {
	return c.where
}

func StringifyFieldValue(value interface{}) (string, bool) {

	foundFloat, ok := value.(float64)
	if ok {
		stringifiedNumber := strconv.FormatFloat(foundFloat, 'f', -1, 64)
		return stringifiedNumber, true
	} else {
		foundString, ok := value.(string)

		if ok {
			formattedString := "'" + foundString + "'"
			return formattedString, true
		}
	}

	return "", false

}

func GetFormattedBody(req *http.Request) *CrudRequest {
	decoder := json.NewDecoder(req.Body)
	values := make([]FieldValue, 0)
	filters := make([]FieldFilter, 0)

	for {
		var body map[string]interface{}
		err := decoder.Decode(&body)

		if err != nil {
			if err == io.EOF {
				fmt.Println(values)
				fmt.Println(filters)
				return &CrudRequest{filters, values}
			} else {
				log.Fatal(err)
			}
		}

		for key, value := range body {
			formattedValue, succeeded := StringifyFieldValue(value)
			if succeeded {
				values = append(values, FieldValue{key, formattedValue})
			} else {
				fmt.Println(value)

				if key == WHERE_CLAUSE {
					foundMap, ok := value.(map[string]interface{})

					if ok {
						for whereKey, whereValue := range foundMap {
							formattedWhereValue, succeeded := StringifyFieldValue(whereValue)
							if succeeded {
								filters = append(filters, FieldFilter{whereKey, "=", formattedWhereValue})
							} else {
								log.Fatal("this is a type that's not available in where clauses")
							}
						}

					} else {
						log.Fatal("this type hasn't been handled by the response body formatter")
					}

				}

			}

		}

	}

	defer req.Body.Close()

	return &CrudRequest{filters, values}
}

type Route struct {
	model  *Model
	cruder *Cruder
}

func (r *Route) Handler(w http.ResponseWriter, crudType string, values *CrudRequest) {
	fmt.Println(w)
	fmt.Println(crudType)
	fmt.Println(values)

	//io.WriteString(w, "successful send")

	switch crudType {
	case "read":
		r.cruder.read(values)
	case "update":
		r.cruder.update(values)
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

	formattedRequest := GetFormattedBody(r)
	crudType := foundPath[2]
	modelName := foundPath[1]
	route := s.routes[modelName]
	route.Handler(w, crudType, formattedRequest)
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
