package main

import (
	"log"
	"tubefeed/internal/app"
)

var version = "dev"

func main() {
	app := app.Setup(version)
	log.Fatal(app.Run())
}
