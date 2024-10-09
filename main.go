package main

import (
	_ "embed"
	"log"
	"tubefeed/internal/app"
)

func main() {
	app := app.Setup()
	log.Fatal(app.Run())
}
