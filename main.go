package main

import (
	"fmt"
	"tubefeed/pkg/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}
