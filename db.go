package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strings"
)

type DataBase struct {
	db *sql.DB
}

func (d *DataBase) Prepare(statement string) *sql.Stmt {
	stmt, err := d.db.Prepare(statement)
	if err != nil {
		log.Fatal(err)
	}

	return stmt
}

// TODO this doesn't cover parametric types
func GetFormattedColumns(fields []Field) string {
	formattedColumns := make([]string, len(fields))
	for i, field := range fields {
		formattedColumns[i] = field.GetName() + " " + field.GetKind() + " "
	}

	fmt.Println(formattedColumns)

	return strings.Join(formattedColumns, ",")
}

func (d *DataBase) CreateTable(model *Model) {
	fmt.Println(model)
	_, err := d.db.Exec("CREATE TABLE " + model.GetName() + " ( " + GetFormattedColumns(model.GetFields()) + ")")

	if err != nil {
		panic(err)
	}
}

func (d *DataBase) InitTables(models []*Model) {
	for _, model := range models {
		if !d.TableExists(model.GetName()) {
			d.CreateTable(model)
		}
	}
}

func (d *DataBase) TableExists(modelName string) bool {
	db := d.db
	query := "SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = ?"
	rows, err := db.Query(query, modelName)

	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	if rows.Next() {
		return true
	}

	return false
}

func InitDatabase(user string, pw string, domain string, port string, dbName string) *sql.DB {
	db, err := sql.Open("mysql", user+":"+pw+"@tcp("+domain+":"+port+")/"+dbName)

	err = db.Ping()
	if err != nil {
		log.Fatal(err.Error())
	}

	_, err = db.Exec("CREATE DATABASE " + dbName)

	if err != nil {
		// db may already exist

		_, err = db.Exec("USE " + dbName)
		if err != nil {
			panic(err)
		}

		err = db.Ping()
		if err != nil {
			panic(err.Error())
		}

	} else {
		_, err = db.Exec("USE " + dbName)
		if err != nil {
			panic(err)
		}
	}

	return db
}
