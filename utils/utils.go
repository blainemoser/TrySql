package utils

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
	if limit < 10 {
		limit = 10
	}
	if len(*input) <= limit {
		return
	}
	*input = (*input)[0:limit-3] + "..."
}

func HandelPanic(t *testing.T) {
	r := recover()
	if r != nil {
		t.Error(r)
	}
}

// MakeBearer creates a bearer token
func MakePass() (string, []byte) {
	rand.Seed(time.Now().Unix())
	bytes := randBytes(32)
	return string(bytes), bytes
}

func randBytes(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return b
}
