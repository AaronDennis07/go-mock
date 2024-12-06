package main

import (
	"log"

	gomock "github.com/AaronDennis07/go-mock/cmd/go-mock"
)

func main() {
	if err := gomock.Run(); err != nil {
		log.Fatal(err)
	}
}
