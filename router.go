package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

const (
	WHERE_CLAUSE    = "where"
	CREATE          = "create"
	READ            = "read"
	UPDATE          = "update"
	DELETE          = "delete"
	VIEW_REGEXP     = "(.*).html"
	ASSETS_DIR_NAME = "public"
)

var VIEW_PATTERN = regexp.MustCompile(VIEW_REGEXP)
var allCrudTypes = []string{CREATE, READ, UPDATE, DELETE}

// contains query details
type CrudRequest struct {
	where []FieldFilter
	main  []FieldValue
}

// contains data used to create records, or update records that match
// a where clause/FieldFilter
func (c *CrudRequest) GetValues() []FieldValue {
	return c.main
}

// contains data used to specify to which records a main clause
// or a crud type should be applied
func (c *CrudRequest) GetFilters() []FieldFilter {
	return c.where
}

// takes json body and formats as a CrudRequest
func GetFormattedBody(req *http.Request) *CrudRequest {
	decoder := json.NewDecoder(req.Body)
	values := make([]FieldValue, 0)
	filters := make([]FieldFilter, 0)

	for {
		var body map[string]interface{}
		err := decoder.Decode(&body)

		if err != nil {
			if err == io.EOF {

				sort.Sort(ByFieldFilterName(filters))
				sort.Sort(ByFieldValueName(values))

				return &CrudRequest{filters, values}
			} else {
				log.Fatal(err)
			}
		}

		for key, value := range body {

			if key == WHERE_CLAUSE {
				foundMap, ok := value.(map[string]interface{})

				if ok {
					for whereKey, whereValue := range foundMap {
						filters = append(filters, FieldFilter{whereKey, "=", whereValue})
					}

				} else {
					log.Fatal("this type hasn't been handled by the response body formatter")
				}

			} else {
				values = append(values, FieldValue{key, value})
			}

		}

	}

	defer req.Body.Close()

	sort.Sort(ByFieldFilterName(filters))
	sort.Sort(ByFieldValueName(values))

	return &CrudRequest{filters, values}
}

// links together a data model and a Cruder, which will handles
// creating, reading, updating, and deleteing records for which this
// model is a schema
type Route struct {
	model  *Model
	cruder *Cruder
}

// delegates requests based on crud type
func (r *Route) Handler(w http.ResponseWriter, crudType string, values *CrudRequest) {

	switch crudType {
	case READ:
		foundRows := r.cruder.read(values)
		encodedRows, err := json.Marshal(foundRows)
		if err != nil {
			log.Fatal(err)
		}
		w.Write(encodedRows)

	case UPDATE:
		r.cruder.update(values)

	case CREATE:
		r.cruder.create(values)

	case DELETE:
		r.cruder.remove(values)

	default:
		return
	}

}

// delegates all CRUD requests to different routes
// informed by its pathValidator
type Router struct {
	routes        map[string]*Route
	pathValidator *regexp.Regexp
	viewValidator *regexp.Regexp
}

// generates routeSpec.txt from all Routes in the Router
func (s *Router) GenerateRouteSpec() {
	spec := ""
	var model *Model
	var modelName string
	var base string

	for _, route := range s.routes {
		model = route.model
		modelName = model.GetName()
		spec += "Routes for the " + modelName + " model\n\n"
		base = "POST /" + modelName + "/"

		for _, ct := range allCrudTypes {
			spec += base + ct + ", payload: " + GetExamplePayload(ct, model) + "\n"
		}

		spec += "\n"
	}

	allViews := FindTemplates()
	for _, view := range allViews {
		spec += "GET /" + view + ", the " + view + " view\n"
	}

	fileContent := []byte(spec)
	err := ioutil.WriteFile("routeSpec.txt", fileContent, 0644)

	if err != nil {
		log.Fatal(err)
	}

}

func GetExamplePayload(crudType string, model *Model) string {
	switch crudType {
	case CREATE:
		return "{" + model.GetFormattedColumns() + " }"
	case READ:
		return "{ where: { " + model.GetFormattedColumns() + " } }"
	case UPDATE:
		return "{" + model.GetFormattedColumns() + ", where: { " + model.GetFormattedColumns() + " } }"
	case DELETE:
		return "{ where: { " + model.GetFormattedColumns() + " } }"
	default:
		log.Fatal(crudType + " is not a crud type")
	}

	return ""
}

// sends a request to a certain route based on the url
func (s *Router) DelegateRequest(w http.ResponseWriter, r *http.Request) {
	foundPath := s.pathValidator.FindStringSubmatch(r.URL.Path)

	if foundPath == nil {
		foundView := s.viewValidator.FindStringSubmatch(r.URL.Path)

		if foundView == nil {
			http.NotFound(w, r)
			return
		}

		// serve view
		view := "./public/" + foundView[1] + ".html"
		http.ServeFile(w, r, view)
	}

	formattedRequest := GetFormattedBody(r)
	crudType := foundPath[2]
	modelName := foundPath[1]
	route := s.routes[modelName]
	route.Handler(w, crudType, formattedRequest)
}

// generates the Router's pathValidator from the accepted crud types
// and existing Models
func GenerateRouteValidator(models []*Model) *regexp.Regexp {
	baseChecker := make([]string, len(models))
	crudTypeChecker := "(" + strings.Join(allCrudTypes, "|") + ")"

	for i, model := range models {
		baseChecker[i] = model.GetName()
	}

	regexpContent := "^/(" + strings.Join(baseChecker, "|") + ")/" + crudTypeChecker + "$"
	return regexp.MustCompile(regexpContent)
}

// returns names of templates found in public directory
func FindTemplates() []string {
	files, err := ioutil.ReadDir(ASSETS_DIR_NAME)
	if err != nil {
		log.Fatal(err)
	}

	views := make([]string, len(files))
	for i, file := range files {
		foundViewName := VIEW_PATTERN.FindStringSubmatch(file.Name())

		if foundViewName == nil {
			log.Fatal("all views must have names and must be html templates")
		}

		views[i] = foundViewName[1]
	}

	return views
}

// generates regexp validator for html templates found
// in public directory
func GenerateViewValidator() *regexp.Regexp {
	templateNames := FindTemplates()
	return regexp.MustCompile("^/(" + strings.Join(templateNames, "|") + ")$")
}

// initializes the Router from the database and generated Models
// attaches generated Cruders
func InitRouter(db *DataBase, models []*Model) *Router {

	routes := make(map[string]*Route, 0)
	cruders := InitCruders(db, models)

	for _, model := range models {
		modelName := model.GetName()
		routes[modelName] = &Route{model, cruders[modelName]}
	}

	return &Router{
		routes,
		GenerateRouteValidator(models),
		GenerateViewValidator(),
	}
}
