package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/blainemoser/TrySql/configs"
	"github.com/blainemoser/TrySql/trysql"
)

var ts *trysql.TrySql

func main() {
	initialise()
	err := ts.Run()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	err = ts.TearDown()
	if err != nil {
		log.Fatal(err.Error())
		return
	}
}

func initialise() {
	var err error
	args := os.Args[1:]
	confs, err := getInputs(args)
	if err != nil {
		log.Fatal(err.Error())
	}
	ts, err = trysql.Init(getProcessOwner(), confs.GetMysqlVersion())
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("found " + string(ts.DockerVersion()))
	err = ts.Provision()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = ts.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getProcessOwner() string {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(stdout)
}

func getInputs(args []string) (*configs.Configs, error) {
	return configs.New(args)
}
