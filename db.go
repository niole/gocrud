package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type DataBase struct {
	db *sql.DB
}

func (d *DataBase) Prepare(statement string) *sql.Stmt {
	stmt, err := d.db.Prepare(statement)
	if err != nil {
		fmt.Println("prepare fail")
		log.Fatal(err)
	}

	return stmt
}

func (d *DataBase) CreateTable(model *Model) {
	query := "CREATE TABLE IF NOT EXISTS " + model.GetName() + "( " + model.GetFormattedColumnsWithTypes() + ")"
	_, err := d.db.Exec(query)

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
