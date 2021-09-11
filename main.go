package main

import (
	"log"

	"github.com/blainemoser/TrySql/shell"
	"github.com/blainemoser/TrySql/trysql"
)

var ts *trysql.TrySql

func main() {
	ts = trysql.Initialise()
	c := shell.New(ts)
	c.Start(false)
	err := ts.TearDown()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}
