// main starts the gomock server and logs any fatal errors during startup.
package main

import (
	"log"

	gomock "github.com/AaronDennis07/go-mock/cmd/go-mock"
)

// main starts the gomock server and logs any fatal errors during startup.
func main() {
	if err := gomock.Run(); err != nil {
		log.Fatal(err)
	}
}
