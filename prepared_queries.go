package main

import (
	"database/sql"
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

type ByName []Field

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].name < a[j].name }

// no duplicate fields, all valid fields, no unspecified
// fields
func (m *Model) AreValidFields(fields []Field) bool {
	seenFields := make(map[string]bool)
	for _, field := range fields {
		if !m.fieldsSpec[field.name] || seenFields[field.name] {
			return false
		}
		seenFields[field.name] = true
	}

	return true
}

type BaseCruder interface {
	create([]Field)
	read(string) []Result
	update([]Field, string)
	remove(string)
}

type Cruder struct {
	db              *DataBase
	model           *Model
	createStatement *Stmt
	readStatement   *Stmt
	updateStatement *Stmt
	removeStatement *Stmt
}

func (c *Cruder) SortFields(fields []Field) []Field {
	return sort.Sort(ByName(fields))
}

func (c *Cruder) create(fields []Field) {
	if c.model.AreValidFields(fields) {
		// create new, grab prepared query

		sortedFields := c.SortFields(fields)
		_, err := c.createStatement.Exec(Map(sortedFields, func(field Field) { return field.value })...)

		if err != nil {
			log.Fatal(err)
		}
	}
}

func (c *Cruder) read(id string) []interface{} {
	rows, err := c.readStatement.Query(id)

	defer rows.Close()

	if err != nil {
		err = rows.Err()
		log.Fatal(err)
	}

	allRows := make([]interface{}, 0)
	for rows.Next() {
		var id int
		var name string
		err = rows.Scan(&id, &name)
		append(allRows, Result{id, name})

		if err != nil {
			log.Fatal(err)
		}
	}

	return allRows
}

func (c *Cruder) update(fields []Field, id string) {
	_, err := c.updateStatement.Exec(id)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cruder) remove(id string) {
	_, err := c.removeStatement(id)
	if err != nil {
		log.Fatal(err)
	}
}

func PrepareCreateStatement(db *DataBase, modelName string) *Stmt {
	return db.Prepare("INSERT INTO " + modelName + "VALUES (?)")
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

func InitCruders(db *sql.DB, models []*Model) (cruders map[string]*Cruder) {
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
