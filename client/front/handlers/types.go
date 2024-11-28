package handlers

type data struct {
	User         string
	Message      string
	Name         string
	RegularChats []string
	SecretChats  []string
	Messages     []message
}

type message struct {
	Author  string
	Message string
}
