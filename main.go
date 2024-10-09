package main

import (
	"log"
	"tubefeed/internal/app"
)

func main() {
	app := app.Setup()
	log.Fatal(app.Run())
}
