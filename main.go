package main

import (
	"log"

	"github.com/blainemoser/TrySql/shell"
	"github.com/blainemoser/TrySql/trysql"
)

func main() {
	ts, err := trysql.Initialise()
	if err != nil {
		log.Fatal(err.Error())
	}
	c := shell.New(ts)
	c.Start(false)
	err = ts.TearDown()
	if err != nil {
		log.Fatal(err.Error())
	}
}
