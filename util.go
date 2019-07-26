package go_blink_util

import (
	"log"
	"os"
)

// CheckErr check the error
func CheckErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

// just ignore everything
func Ignore(any ...interface{}) {}

// Exists is file exist
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
