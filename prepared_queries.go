package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type BaseCruder interface {
	create([]FieldValue)
	read([]FieldValue) []interface{}
	update([]FieldValue, []FieldValue)
	remove([]FieldValue)
}

type Cruder struct {
	db              *DataBase
	model           *Model
	createStatement *sql.Stmt
	readStatement   *sql.Stmt
	updateStatement *sql.Stmt
	removeStatement *sql.Stmt
}

func (c *Cruder) ValidateCreateStatement(values []FieldValue) bool {
	totalUniqueValues := 0
	foundValues := make(map[string]bool, 0)

	for _, value := range values {
		if foundValues[value.Name] {
			return false
		} else {
			totalUniqueValues += 1
			foundValues[value.Name] = true
		}
	}

	return totalUniqueValues == len(c.model.GetFields())
}

// takes specified values and executes a prepared create statement
func (c *Cruder) create(values []FieldValue) {
	if c.ValidateCreateStatement(values) {
		formattedFields := make([]string, len(values))
		formattedValues := make([]string, len(values))

		for i, value := range values {
			formattedFields[i] = value.Name
			formattedValues[i] = value.Value
		}

		columns := strings.Join(formattedFields, ",")
		allValues := strings.Join(formattedValues, ",")

		_, err := c.createStatement.Exec(columns, allValues)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func (c *Cruder) read(values []FieldValue) []interface{} {
	whereClause := make([]string, len(values))

	for i, value := range values {
		whereClause[i] = value.Name + "=" + value.Value
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

func (c *Cruder) update(values []FieldValue, newValues []FieldValue) {
	setClause := make([]string, len(newValues))
	whereClause := make([]string, len(values))

	for i, value := range newValues {
		setClause[i] = value.Name + "=" + value.Value
	}

	for i, value := range values {
		whereClause[i] = value.Name + "=" + value.Value
	}

	set := strings.Join(setClause, ",")
	where := strings.Join(whereClause, ",")

	_, err := c.updateStatement.Exec(set, where)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cruder) remove(values []FieldValue) {
	whereClause := make([]string, len(values))

	for i, value := range values {
		whereClause[i] = value.Name + "=" + value.Value
	}

	statement := strings.Join(whereClause, ",")
	_, err := c.removeStatement.Exec(statement)
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

func PrepareUpdateStatement(db *DataBase, model *Model) *sql.Stmt {
	fields := model.GetFields()

	setClause := make([]string, len(fields))
	for i, field := range fields {
		setClause[i] = field.GetName() + "=?"
	}

	formattedSetClause := strings.Join(setClause, ",")
	baseQuery := `UPDATE ` + model.GetName() + `
		SET ` + formattedSetClause + `
		WHERE id=?
	`

	return db.Prepare(baseQuery)
}

func PrepareRemoveStatement(db *DataBase, modelName string) *sql.Stmt {
	return db.Prepare("DELETE FROM " + modelName + " WHERE ?")
}

func InitCruders(db *DataBase, models []*Model) (cruders map[string]*Cruder) {
	for _, model := range models {
		modelName := model.GetName()
		fmt.Println("init")

		cruders[modelName] = &Cruder{
			db,
			model,
			PrepareCreateStatement(db, model),
			PrepareReadStatement(db, modelName),
			PrepareUpdateStatement(db, model),
			PrepareRemoveStatement(db, modelName),
		}
	}
	return
}
