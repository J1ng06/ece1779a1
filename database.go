package main

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const ConnectionString = "host=localhost dbname=ece1779 port=5432 sslmode=disable"

func NewConnection() (db *gorm.DB, err error) {

	db = new(gorm.DB)
	db, err = gorm.Open("postgres", ConnectionString)
	if err != nil {
		log.Fatal("could not connect to db; fatal db", err)
		return nil, err
	}
	db.Exec("SET SEARCH_PATH to a1")
	return db, nil

}
