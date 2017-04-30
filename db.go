package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql",
		"root:root@tcp(127.0.0.1:3307)/mysql")

	if err != nil {
		fmt.Println("failed connection")
		panic(err.Error())
	}

	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("failed ping")
		panic(err.Error())
	}

}
