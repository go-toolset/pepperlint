package pepperlint

import (
	"log"
	"os"
)

// Log will be used to log only if PURIFY_DEBUG is set
var Log func(format string, args ...interface{})

func init() {
	if v := os.Getenv("PURIFY_DEBUG"); len(v) != 0 {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		Log = log.Printf
	} else {
		Log = func(format string, args ...interface{}) {}
	}
}
