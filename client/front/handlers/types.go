package handlers

type data struct {
	User         string
	Message      string
	Name         string
	Alert        string
	RegularChats []string
	SecretChats  []string
	Messages     []message
}

type message struct {
	Message string

	Type   string
	Author string
}
