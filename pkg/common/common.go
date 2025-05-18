package common

import (
	"log"
)

func EnsureNotError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
