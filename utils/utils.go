package utils

import (
	"fmt"
	"strings"
)

func GetErrors(errs []error) error {
	var errStrings []string
	if len(errs) > 0 {
		for _, e := range errs {
			errStrings = append(errStrings, e.Error())
		}
		return fmt.Errorf(strings.Join(errStrings, "; "))
	}
	return nil
}
