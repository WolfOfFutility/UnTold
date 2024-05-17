package main

import (
	"log"
)

type testParams struct {
	keyFilePath *string
	keyContent *[]byte
}

func paramFire(params testParams) {
	if params.keyFilePath != nil {
		log.Println("Hello!")
	}
}