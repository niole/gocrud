package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

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

	return models
}

func main() {
	baseModels := IngestJSON()
	dataBase := InitDatabase("root", "root", "127.0.0.1", "3307", "mysql")

	db := &DataBase{dataBase}
	db.InitTables(baseModels)
	router := InitRouter(db, baseModels)
	http.HandleFunc("*", router.DelegateRequest)
	http.ListenAndServe(":8080", nil)
}
