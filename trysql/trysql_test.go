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
	result, err := tsql.listContainers(false)
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

func TestPassword(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result := tsql.CurrentPassword()
	if len(result) != 32 {
		t.Errorf("expected password to be 32 characters, got %d", len(result))
	}
	temp := tsql.DockerTempPassword()
	if result == temp {
		t.Errorf("expected temp password to be different from new password")
	}
}

func TestGetContainerDetails(t *testing.T) {
	defer utils.HandelPanic(t)
	tInit()
	result := tsql.GetContainerDetails(false)
	if !strings.Contains(strings.ToLower(result), "trysql") {
		t.Errorf("expected to find the container name 'TrySql', got '%s'", result)
	}
	result = tsql.GetContainerDetails(true)
	expects := map[string]bool{
		"not found": true,
		"something went wrong while trying to get the container's details": true,
	}
	if expects[result] {
		t.Errorf(result)
	}
}

func TestMysqlArgs(t *testing.T) {
	defer utils.HandelPanic(t)
	result := tsql.mysqlArgs("QUERY")
	if !strings.Contains(result, "--execute=\"QUERY\"") {
		t.Errorf("expected raw string to contain '--execute=\"QUERY\"', got %s", result)
	}
}

func TestQuery(t *testing.T) {
	defer utils.HandelPanic(t)
	result, err := tsql.Query("SHOW VARIABLES LIKE 'max_connections'", true)
	if err != nil {
		t.Error(err)
	}
	if !strings.Contains(result, "Variable_name") {
		t.Errorf("expected result to contain 'Variable_name', got '%s'", result)
	}
}

func TestContainerRunning(t *testing.T) {
	defer utils.HandelPanic(t)
	result, err := tsql.containerRunning()
	if err != nil {
		t.Error(err)
	}
	if !result {
		t.Errorf("expected container to be running")
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
