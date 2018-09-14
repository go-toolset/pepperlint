package pepperlint

import (
	"log"
	"os"
)

// Log will be used to log only if PEPPERLINT_DEBUG is set
var Log func(format string, args ...interface{})

func init() {
	if v := os.Getenv("PEPPERLINT_DEBUG"); len(v) != 0 {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		Log = log.Printf
	} else {
		Log = func(format string, args ...interface{}) {}
	}
}
