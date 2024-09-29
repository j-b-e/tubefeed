package config

import (
	"database/sql"
	"fmt"
)

const (
	ListenPort = 8091
	AudioPath  = "./audio/"
	DbPath     = "./config/tubefeed.db"
	Hostname   = "luchs"
)

var ExternalURL = fmt.Sprintf("%s:%d", Hostname, ListenPort)
var Db *sql.DB
