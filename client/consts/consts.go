package consts

import "messengerClient/types"

var EventListeners types.EventsType
var Recievers types.RecievedType

func AddListener(user string, eventID string, listener chan int) {
	EventListeners.Mu.Lock()
	EventListeners.Events[user][eventID] = listener
	EventListeners.Mu.Unlock()
}

func RemoveListener(user string, eventID string) {
	EventListeners.Mu.Lock()
	listener := EventListeners.Events[user][eventID]
	if listener != nil {
		close(listener)
		delete(EventListeners.Events[user], eventID)
	}
	EventListeners.Mu.Unlock()
}

const (
	LocalHost = "localhost:"
	LocalIP   = "127.0.0.1:"

	EncriptionNo      = "no"
	EncriptionMagenta = "magenta"
	EncriptionRC6     = "rc6"
)
