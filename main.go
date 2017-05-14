package main

import (
	"net/http"
)

func main() {
	baseModels := IngestJSON()
	dataBase := InitDatabase("root", "root", "127.0.0.1", "3307", "mysql")

	db := &DataBase{dataBase}
	db.InitTables(baseModels)
	router := InitRouter(db, baseModels)
	http.HandleFunc("/", router.DelegateRequest)
	http.ListenAndServe(":8080", nil)
}
