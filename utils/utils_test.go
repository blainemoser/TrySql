package utils

import (
	"fmt"
	"testing"
)

func TestGetProcessOwner(t *testing.T) {
	defer HandelPanic(t)
	GetProcessOwner()
}

func TestGetErrors(t *testing.T) {
	defer HandelPanic(t)
	var errs []error
	errs = append(errs, fmt.Errorf("error one test"))
	errs = append(errs, fmt.Errorf("error two test"))
	result := GetErrors(errs)
	expects := "error one test; error two test"
	if result.Error() != expects {
		t.Error(fmt.Errorf("expected error to be %s, got %s", expects, result.Error()))
	}
}

func TestTruncString(t *testing.T) {
	defer HandelPanic(t)
	test1 := "under ten"
	TruncString(&test1, 1)
	if test1 != "under ten" {
		t.Error(fmt.Errorf("expected string to be 'under ten', got '%s'", test1))
	}
	test2 := "over ten characters long"
	TruncString(&test2, 15)
	if test2 != "over ten cha..." {
		t.Error(fmt.Errorf("expected string to be 'over ten cha...', got '%s'", test2))
	}
}

func TestHandlePanic(t *testing.T) {
	nt := &testing.T{}
	defer HandelPanic(nt)
	triggerPanic(nt)
}

func triggerPanic(nt *testing.T) {
	panic(fmt.Errorf("test panic"))
}
