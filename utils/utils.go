package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func GetProcessOwner() string {
	stdout, err := exec.Command("ps", "-o", "user=", "-p", strconv.Itoa(os.Getpid())).Output()
	if err != nil {
		log.Fatal(err)
	}
	return string(stdout)
}

func GetErrors(errs []error) error {
	var errStrings []string
	if len(errs) > 0 {
		for _, e := range errs {
			if e != nil {
				errStrings = append(errStrings, e.Error())
			}
		}
	}
	if len(errStrings) > 0 {
		return fmt.Errorf(strings.Join(errStrings, "; "))
	}
	return nil
}

func TruncString(input *string, limit int) {
	if len(*input) <= limit {
		return
	}
	*input = (*input)[0:limit-3] + "..."
}

func StringInList(needle string, haystack []string) bool {
	for _, h := range haystack {
		if needle == h {
			return true
		}
	}
	return false
}

func DelineTab(input *string) {
	*input = strings.ReplaceAll(*input, "\t", " ")
	*input = strings.ReplaceAll(*input, "\n", " ")
}
