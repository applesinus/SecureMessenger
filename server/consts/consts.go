package consts

import (
	"log"
)

const (
	RabbitmqAPI = "http://localhost:15672/api"
	Vhost       = "%2F"
)

var (
	RabbitmqUser     = "admin"
	RabbitmqPassword = "admin"
)

func LogIfError(err error, msg string) bool {
	if err != nil {
		log.Printf("%s: %s", msg, err)
		return true
	}
	return false
}
