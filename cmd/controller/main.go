package main

import (
	"fmt"

	"github.com/pupload/pupload/internal/controller"
)

func main() {
	if err := controller.Run(); err != nil {
		fmt.Printf("error running controller: %s", err)
	}
}
