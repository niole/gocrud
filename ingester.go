package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
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
	dataBase := InitDataBase("root", "root", "127.0.0.1", "3306", "nioledb")

	db := &DataBase{dataBase}
	db.InitTables(baseModels)
	cruders := InitCruders(db, baseModels)
}
