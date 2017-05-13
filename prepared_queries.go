package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"sort"
)

func Map(baseArray []interface{}, F func(interface{}) interface{}) []interface{} {
	newArray := make([]interface{}, len(baseArray))

	for i, elt := range baseArray {
		newArray[i] = F(elt)
	}

	return newArray
}

type BaseCruder interface {
	create([]FieldValue)
	read([]FieldValue) []interface{}
	update([]FieldValue, []FieldValue)
	remove([]FieldValue)
}

type Cruder struct {
	db              *DataBase
	model           *Model
	createStatement *Stmt
	readStatement   *Stmt
	updateStatement *Stmt
	removeStatement *Stmt
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

	if totalUniqueValues != len(model.Fields) {
		return false
	}

	return true
}

// takes specified values and executes a prepared create statement
func (c *Cruder) create(values []FieldValue) {
	if ValidateCreateStatement(values) {
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

func (c *Cruder) read(values []FieldValues) []interface{} {
	whereClause := make([]string, len(values))

	for i, value := range values {
		whereClause[i] = value.Name + "=" + value.Value
	}

	statement := strings.Join(whereClause, ",")
	rows, err := c.readStatement.Query("*", statement)

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

		newJson, err := json.Marshall(all)

		if err != nil {
			log.Fatal(err)
		}

		append(allRows, Result{id, name})
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
	_, err := c.removeStatement(statement)
	if err != nil {
		log.Fatal(err)
	}
}

func PrepareCreateStatement(db *DataBase, modelName string) *Stmt {
	return db.Prepare("INSERT INTO " + modelName + " (?) VALUES (?)")
}

func PrepareReadStatement(db *DataBase, modelName string) *Stmt {
	return db.Prepare("SELECT ? FROM " + modelName + " WHERE ?")
}

func PrepareUpdateStatement(db *DataBase, modelName string) *Stmt {
	return db.Prepare("UPDATE " + modelName + " SET ? WHERE ?")
}

func PrepareRemoveStatement(db *DataBase, modelName string) *Stmt {
	return db.Prepare("DELETE FROM " + modelName + " WHERE ?")
}

func InitCruders(db *DataBase, models []*Model) (cruders map[string]*Cruder) {
	for _, model := range models {
		modelName := model.GetName()

		cruders[model.GetName()] = &Cruder{
			db,
			model,
			PrepareCreateStatement(db, modelName),
			PrepareReadStatement(db, modelName),
			PrepareUpdateStatement(db, modelName),
			PrepareRemoveStatement(db, modelName),
		}
	}

}
