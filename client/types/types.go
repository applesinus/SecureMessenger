package types

import "sync"

type Message struct {
	Id      string
	Message string

	Type   string
	Author string
}

type EventsType struct {
	Mu     *sync.Mutex
	Events map[string]map[string]chan int
}

type RecievedType struct {
	Mu     *sync.Mutex
	Events map[string]map[string]chan []byte
}

type Chats struct {
	Mu    *sync.Mutex
	Chats map[string]*ChatType
}

type ChatType struct {
	Id         string    `json:"id"`
	Reciever   string    `json:"reciever"`
	Encryption string    `json:"encryption"`
	Messages   []Message `json:"messages"`
}

type Data struct {
	User         string
	Message      string
	Name         string
	Alert        string
	RegularChats [][]string
	SecretChats  [][]string
	Messages     []Message
	Listeners    []string
}
