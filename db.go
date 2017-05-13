package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type DataBase struct {
	db *sql.db
}

func (d *DataBase) Prepare(statement string) *Stmt {
	stmt, err := d.db.Prepare(statement)
	if err != nil {
		log.Fatal(err)
	}

	return stmt
}

// TODO this doesn't cover parametric types
func GetFormattedColumns(fields []Field) (formattedColumns string) {
	for _, field := range fields {
		formattedColumns += field.GetName() + " " + field.GetKind() + " "
	}
}

func (d *DataBase) CreateTable(model *Model) {
	_, err = d.db.Exec("CREATE TABLE " + model.GetName() + " ( " + GetFormattedColumns(model.fields) + ")")

	if err != nil {
		panic(err)
	}
}

func (d *DataBase) InitTables(models []*Model) {
	for _, model := range models {
		d.CreateTable(model)
	}
}

func InitDatabase(user string, pw string, domain string, port string, dbName string) *sql.DB {
	db, err := sql.Open("mysql", user+":"+pw+"@tcp("+domain+":"+port+")/"+dbName)

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec("USE " + dbName)
	if err != nil {
		panic(err)
	}

	return db
}
