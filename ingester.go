package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
)

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByName []Field

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

// sort fields by name
func SortAllModelFields(models []*Model) []*Model {
	for _, model := range models {
		fmt.Println(model.Fields)
		sort.Sort(ByName(model.Fields))
		fmt.Println(model.Fields)
	}

	return models
}

func IngestJSON() []*Model {
	output, err := ioutil.ReadFile("./models.json")

	if err != nil {
		log.Fatal(err)
	}

	var models []*Model

	err = json.Unmarshal(output, &models)
	if err != nil {
		log.Fatal(err)
	}

	return SortAllModelFields(models)
}
