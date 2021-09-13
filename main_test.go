package main

import (
	"os"
	"testing"

	"github.com/blainemoser/TrySql/shell"
	"github.com/blainemoser/TrySql/utils"
)

var suite *shell.TestSuiteTS

func TestMain(m *testing.M) {
	var err error
	suite, err = shell.InitialiseTestSuite()
	if err != nil {
		panic(err)
	}
	suite.Start()
	code := m.Run()
	err = suite.Stop()
	if err != nil {
		panic(err)
	}
	os.Exit(code)
}

func TestRun(t *testing.T) {
	defer utils.HandelPanic(t)
}
