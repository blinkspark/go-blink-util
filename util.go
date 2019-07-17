package go_blink_util

import "log"

// CheckErr check the error
func CheckErr(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

// just ignore everything
func Ignore(any ...interface{}) {}
