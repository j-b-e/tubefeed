package main

import (
	"fmt"
	"tubefeed/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Println(err)
	}
}
