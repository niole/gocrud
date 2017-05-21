package main

import (
	"database/sql"
	"log"
	"sort"
	"strings"
)

type BaseCruder interface {
	create(*CrudRequest)
	read(*CrudRequest) []map[string]interface{}
	update(*CrudRequest)
	remove(*CrudRequest)
}

type Cruder struct {
	db                 *DataBase
	model              *Model
	createStatement    *sql.Stmt
	preparedStatements map[string]*sql.Stmt
}

func (c *Cruder) generateStatementKey(crudType string, request *CrudRequest) string {
	filters := request.GetFilters()
	values := request.GetValues()

	filterNames := make([]string, len(filters))
	for i, filter := range filters {
		filterNames[i] = filter.GetName()
	}

	valueNames := make([]string, len(values))
	for i, value := range values {
		valueNames[i] = value.GetName()
	}

	return crudType + strings.Join(filterNames, "|") + strings.Join(valueNames, "|")
}

func (c *Cruder) generatePreparedStatement(crudType string, request *CrudRequest) *sql.Stmt {
	var err error
	var stmt *sql.Stmt

	switch crudType {
	case READ:
		stmt, err = c.db.db.Prepare(getReadBaseQuery(c.model.GetName(), request))
	case CREATE:
		log.Fatal("generating prepared statments for create calls is unecessary")
	case UPDATE:
		stmt, err = c.db.db.Prepare(getUpdateBaseQuery(c.model.GetName(), request))
	case DELETE:
		stmt, err = c.db.db.Prepare(getRemoveBaseQuery(c.model.GetName(), request))
	default:
		log.Fatal(crudType + " is not a real crudType")
	}

	if err != nil {
		log.Fatal(err)
	}

	return stmt
}

// returns cached prepared statement or creates new, caches, and returns
func (c *Cruder) getPreparedStatement(crudType string, request *CrudRequest) *sql.Stmt {
	key := c.generateStatementKey(crudType, request)
	foundStatement := c.preparedStatements[key]

	if foundStatement == nil {
		foundStatement = c.generatePreparedStatement(crudType, request)
		c.preparedStatements[key] = foundStatement
	}

	return foundStatement
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

func (c *Cruder) read(request *CrudRequest) []map[string]interface{} {
	values := request.GetFilters()
	totalColumns := len(c.model.GetFields()) + 1 // add 1 to include ids
	filterValues := make([]interface{}, len(values))

	for i, value := range values {
		filterValues[i] = value.GetValue()
	}

	statement := c.getPreparedStatement(READ, request)
	rows, err := statement.Query(filterValues...)

	defer rows.Close()

	if err != nil {
		err = rows.Err()
		log.Fatal(err)
	}

	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	allRows := make([]map[string]interface{}, 0)

	for rows.Next() {
		all := make([]interface{}, totalColumns)
		allContent := make([]string, totalColumns)

		for i, _ := range allContent {
			all[i] = &allContent[i]
		}

		err = rows.Scan(all...)
		if err != nil {
			log.Fatal(err)
		}

		formattedJSON := make(map[string]interface{})

		for i, colName := range cols {
			formattedJSON[colName] = all[i]
		}

		allRows = append(allRows, formattedJSON)
	}

	return allRows
}

// TODO validate queries better, user can currently send columns that don't exist
func (c *Cruder) update(request *CrudRequest) {
	newValues := request.GetValues()
	filters := request.GetFilters()

	setClause := make([]interface{}, len(newValues))
	whereClause := make([]interface{}, len(filters))

	for i, value := range newValues {
		setClause[i] = value.Value
	}

	for i, filter := range filters {
		whereClause[i] = filter.Value
	}

	arguments := append(setClause, whereClause...)

	statement := c.getPreparedStatement(UPDATE, request)

	_, err := statement.Exec(arguments...)

	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cruder) remove(request *CrudRequest) {
	values := request.GetFilters()
	whereStatement := make([]interface{}, len(values))

	for i, value := range values {
		whereStatement[i] = value.Value
	}

	statement := c.getPreparedStatement(DELETE, request)

	_, err := statement.Exec(whereStatement...)
	if err != nil {
		log.Fatal(err)
	}
}

func getCreateBaseQuery(model *Model) string {
	modelName := model.GetName()
	formattedColumns := model.GetFormattedColumns()
	interpValuePlaceholders := make([]string, len(model.GetFields()))

	for i, _ := range interpValuePlaceholders {
		interpValuePlaceholders[i] = "?"
	}

	placeholders := strings.Join(interpValuePlaceholders, ",")

	return `
		INSERT INTO ` + modelName + `(` + formattedColumns + `)
		VALUES(` + placeholders + `)
	`
}

func getRemoveBaseQuery(modelName string, request *CrudRequest) string {
	filters := request.GetFilters()
	wherePlaceholder := make([]string, len(filters))

	for i, filter := range filters {
		wherePlaceholder[i] = filter.GetSerializedFilter()
	}

	placeholder := strings.Join(wherePlaceholder, ",")

	return "DELETE FROM " + modelName + " WHERE " + placeholder
}

func getUpdateBaseQuery(modelName string, request *CrudRequest) string {
	newValues := request.GetValues()
	filters := request.GetFilters()

	setPlaceholders := make([]string, len(newValues))
	wherePlaceholders := make([]string, len(filters))

	for i, value := range newValues {
		setPlaceholders[i] = value.GetName() + "=?"
	}

	for i, value := range filters {
		wherePlaceholders[i] = value.GetSerializedFilter()
	}

	setPlaceholder := strings.Join(setPlaceholders, ",")
	wherePlaceholder := strings.Join(wherePlaceholders, ",")

	return `UPDATE ` + modelName + `
		SET ` + setPlaceholder + `
		WHERE ` + wherePlaceholder + `
	`
}

func getReadBaseQuery(modelName string, request *CrudRequest) string {
	filters := request.GetFilters()
	whereClause := make([]string, len(filters))

	for i, filter := range filters {
		whereClause[i] = filter.GetSerializedFilter()
	}

	statement := strings.Join(whereClause, ",")

	return `
		SELECT * FROM ` + modelName + ` WHERE ` + statement + `
	`
}

func PrepareCreateStatement(db *DataBase, model *Model) *sql.Stmt {
	baseQuery := getCreateBaseQuery(model)

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
			make(map[string]*sql.Stmt),
		}
	}

	return cruders
}
