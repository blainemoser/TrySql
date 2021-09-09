package trysql

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/blainemoser/TrySql/utils"
)

var tsql *TrySql
var torn bool
var initialised bool

func TestMain(m *testing.M) {
	Testing = true
	code := m.Run()
	Testing = false
	if !torn {
		err := tsql.TearDown()
		if err != nil {
			panic(err)
		}
	}
	os.Exit(code)
}

func TestInitialize(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
}

func TestListContainers(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result, err := tsql.ListContainers()
	if err != nil {
		t.Error(err)
	}
	expects := "mysql/mysql-server:latest"
	for _, container := range result {
		if strings.Contains(container, expects) {
			return
		}
	}
	t.Error(fmt.Errorf("expected to find container for '%s'", expects))
}

func TestDockerVersion(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result := tsql.DockerVersion()
	if !strings.Contains(strings.ToLower(result), "docker version") {
		t.Error("expected to find docker version")
	}
	if !strings.Contains(strings.ToLower(result), "build") {
		t.Error("expected to find docker version build")
	}
}

func TestTearDown(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	err := tsql.TearDown()
	if err != nil {
		t.Error(err)
	}
	torn = true
}

func tInit() {
	if !initialised {
		tsql = Initialise()
		initialised = true
	}
}
