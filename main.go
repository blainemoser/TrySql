package main

import (
	"log"

	"github.com/blainemoser/TrySql/shell"
	"github.com/blainemoser/TrySql/trysql"
)

var ts *trysql.TrySql

func main() {
	ts = trysql.Initialise()
	err := ts.Run()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	c := shell.New(ts.Configs)
	c.Start()
	err = ts.TearDown()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}
