package types

import (
	"math/big"
	"sync"
)

type Message struct {
	Id      string `json:"id"`
	Message []byte `json:"message"`
	Iv      []byte `json:"iv"`

	Type   string `json:"type"`
	Author string `json:"author"`
}

type EventsType struct {
	Mu     *sync.Mutex
	Events map[string]map[string]map[string]chan int
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
	Algorithm  string    `json:"algorithm"`
	Padding    string    `json:"padding"`
	Messages   []Message `json:"messages"`
	Keys       Keys      `json:"keys"`
}

type Keys struct {
	Prime             *big.Int `json:"prime"`
	PrimitiveRoot     *big.Int `json:"primitiveRoot"`
	MyPrivateKey      *big.Int `json:"myPrivateKey"`
	RecieverPublicKey *big.Int `json:"recieverPublicKey"`
}

type Data struct {
	User         string
	Message      string
	Name         string
	Alert        string
	RegularChats []string
	SecretChats  []string
	Messages     []Message
	Listeners    []string
}
