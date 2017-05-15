package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
)

type BaseCruder interface {
	create(*CrudRequest)
	read(*CrudRequest) []interface{}
	update(*CrudRequest)
	remove(*CrudRequest)
}

type Cruder struct {
	db              *DataBase
	model           *Model
	createStatement *sql.Stmt
	readStatement   *sql.Stmt
}

// takes specified values and executes a prepared create statement
func (c *Cruder) create(request *CrudRequest) {
	values := request.GetValues()
	modelFields := c.model.GetFields()

	if len(values) == len(modelFields) {
		formattedValues := make([]interface{}, len(values))

		sort.Sort(ByFieldValueName(values))

		for i, value := range values {
			formattedValues[i] = value.Value
		}

		_, err := c.createStatement.Exec(formattedValues...)

		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Fatal("total columns in create statement doesn't match total columns in table")
	}

}

func (c *Cruder) read(request *CrudRequest) []interface{} {
	values := request.GetFilters()
	whereClause := make([]string, len(values))

	for i, value := range values {
		whereClause[i] = value.GetSerializedFilter()
	}

	statement := strings.Join(whereClause, ",")
	rows, err := c.readStatement.Query(statement)

	defer rows.Close()

	if err != nil {
		err = rows.Err()
		log.Fatal(err)
	}

	allRows := make([]interface{}, 0) // this should be returned as JSON
	for rows.Next() {
		var all interface{}
		err = rows.Scan(&all)
		if err != nil {
			log.Fatal(err)
		}

		newJson, err := json.Marshal(all)

		if err != nil {
			log.Fatal(err)
		}

		allRows = append(allRows, newJson)
	}

	return allRows
}

// TODO validate queries better, user can currently send columns that don't exist
func (c *Cruder) update(request *CrudRequest) {
	newValues := request.GetValues()
	values := request.GetFilters()

	setClause := make([]interface{}, len(newValues))
	setPlaceholders := make([]string, len(newValues))

	whereClause := make([]interface{}, len(values))
	wherePlaceholders := make([]string, len(values))

	for i, value := range newValues {
		setClause[i] = value.Value
		setPlaceholders[i] = value.GetName() + "=?"
	}

	for i, value := range values {
		whereClause[i] = value.Value
		wherePlaceholders[i] = value.GetName() + value.GetOp() + "?"
	}

	setPlaceholder := strings.Join(setPlaceholders, ",")
	wherePlaceholder := strings.Join(wherePlaceholders, ",")

	baseQuery := `UPDATE ` + c.model.GetName() + `
		SET ` + setPlaceholder + `
		WHERE ` + wherePlaceholder + `
	`

	arguments := append(setClause, whereClause...)

	_, err := c.db.db.Exec(baseQuery, arguments...)

	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cruder) remove(request *CrudRequest) {
	modelName := c.model.GetName()
	values := request.GetFilters()
	wherePlaceholder := make([]string, len(values))
	whereStatement := make([]interface{}, len(values))

	for i, value := range values {
		wherePlaceholder[i] = value.Name + "=?"
		whereStatement[i] = value.Value
	}

	placeholder := strings.Join(wherePlaceholder, ",")
	baseQuery := "DELETE FROM " + modelName + " WHERE " + placeholder

	_, err := c.db.db.Exec(baseQuery, whereStatement...)
	if err != nil {
		log.Fatal(err)
	}
}

func PrepareCreateStatement(db *DataBase, model *Model) *sql.Stmt {
	formattedColumns := model.GetFormattedColumns()
	interpValuePlaceholders := make([]string, len(model.GetFields()))

	for i, _ := range interpValuePlaceholders {
		interpValuePlaceholders[i] = "?"
	}

	baseQuery := `
		INSERT INTO ` + model.GetName() + `(` + formattedColumns + `)
		VALUES(` + strings.Join(interpValuePlaceholders, ",") + `)
	`
	return db.Prepare(baseQuery)
}

func PrepareReadStatement(db *DataBase, modelName string) *sql.Stmt {
	baseQuery := `
		SELECT * FROM ` + modelName + ` WHERE ?
	`
	return db.Prepare(baseQuery)
}

func InitCruders(db *DataBase, models []*Model) map[string]*Cruder {
	cruders := make(map[string]*Cruder)

	for _, model := range models {
		modelName := model.GetName()

		cruders[modelName] = &Cruder{
			db,
			model,
			PrepareCreateStatement(db, model),
			PrepareReadStatement(db, modelName),
		}
	}

	return cruders
}
