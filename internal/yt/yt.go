package yt

import "fmt"

var Yturl = func(id string) string { return fmt.Sprintf("https://www.youtube.com/watch?v=%s", id) }
