package main

import (
	"tubefeed/internal/app"
)

var version = "dev"

func main() {
	app := app.Setup(version)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
