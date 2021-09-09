package configs

import (
	"fmt"
	"testing"

	"github.com/blainemoser/TrySql/utils"
)

func TestConfigs(t *testing.T) {
	configs, err := New([]string{"--version", "latest", "--buffer-size", "100"})
	if err != nil {
		t.Error(err)
	}
	check(configs, t)
}

func TestEqualsSetting(t *testing.T) {
	configs, err := New([]string{"--version=latest", "--buffer-size=100"})
	if err != nil {
		t.Error(err)
	}
	check(configs, t)
}

func TestNonConfig(t *testing.T) {
	_, err := New([]string{"--tersion", "latest", "--buffer-size", "100"})
	exp := "the tersion argument does not exist"
	if err == nil {
		t.Error(fmt.Errorf("expected error: %s", exp))
		return
	}
	if err.Error() != exp {
		t.Error(fmt.Errorf("expected error to be '%s', got '%s'", exp, err.Error()))
	}
}

func check(configs *Configs, t *testing.T) {
	var errs []error
	version := configs.GetMysqlVersion()
	if version != "latest" {
		errs = append(errs, fmt.Errorf("expected 'mysql-version' to be 'latest'"))
	}
	size := configs.GetBufferSize()
	if size != 100 {
		errs = append(errs, fmt.Errorf("expected 'buffer-size' to be %d", 100))
	}
	err := utils.GetErrors(errs)
	if err != nil {
		t.Error(err)
	}
}
