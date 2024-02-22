package db

import (
	"fmt"
	"log"
	"os"

	dbx "github.com/go-ozzo/ozzo-dbx"
	_ "github.com/lib/pq"
)

var DB *dbx.DB

func Init() {
	uri := fmt.Sprintf("postgres://postgres:%s%s", os.Getenv("DB_PASSWORD"), os.Getenv("DB_URI"))
	db, err := dbx.Open("postgres", uri)
	if err != nil {
		log.Fatal(err)
	}
	DB = db
}
