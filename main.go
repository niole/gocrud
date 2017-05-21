package main

import (
	"net/http"
)

// generates Models from JSON spec defined by user
// initializes the Database, Router, Tables in Database, sets up Route delegation
func main() {
	baseModels := IngestJSON()
	dataBase := InitDatabase("root", "root", "127.0.0.1", "3307", "mysql")

	db := &DataBase{dataBase}
	db.InitTables(baseModels)
	router := InitRouter(db, baseModels)
	http.HandleFunc("/", router.DelegateRequest)
	http.ListenAndServe(":8080", nil)
}
