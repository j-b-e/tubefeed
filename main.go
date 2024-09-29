package main

import (
	"fmt"
	"tubefeed/internal/app"
)

func main() {
	app := app.Setup()
	if err := app.Run(); err != nil {
		fmt.Println(err)
	}
}
