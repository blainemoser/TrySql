package main

import (
	"os"
	"testing"

	"github.com/blainemoser/TrySql/test"
)

var suite *test.TestSuiteTS

func TestMain(m *testing.M) {
	var err error
	suite, err = test.Init()
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

func TestHelp(t *testing.T) {
	defer suite.HandelPanic(t)
	suite.SendHelpSignal()
}

func TestVersion(t *testing.T) {
	defer suite.HandelPanic(t)
	suite.SendVersionSignal()
}
