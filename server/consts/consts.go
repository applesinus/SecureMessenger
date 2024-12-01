package consts

import "log"

const (
	RabbitmqAPI      = "http://localhost:15672/api"
	RabbitmqUser     = "guest"
	RabbitmqPassword = "guest"
	Vhost            = "%2F"
)

func LogIfError(err error, msg string) bool {
	if err != nil {
		log.Printf("%s: %s", msg, err)
		return true
	}
	return false
}
