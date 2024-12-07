package types

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Tags     string `json:"tags"`
}

type Permission struct {
	Configure string `json:"configure"`
	Write     string `json:"write"`
	Read      string `json:"read"`
}

type TopicPermissions struct {
	Exchange string `json:"exchange"`
	Write    string `json:"write"`
	Read     string `json:"read"`
}

type Message struct {
	Id      string `json:"id"`
	Message []byte `json:"message"`

	Type   string `json:"type"`
	Author string `json:"author"`
}
