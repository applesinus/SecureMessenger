package users

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"sync"
)

type UsersType struct {
	rwmu  *sync.RWMutex
	file  *os.File
	users map[string]struct{}
}

var Users UsersType

func LoadUsers(ctx context.Context, wg *sync.WaitGroup) {
	file, err := os.OpenFile("back/users.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("[BACKEND][USERS LOAD] Error opening file: %s", err)
	}

	body, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("[BACKEND][USERS LOAD] Error reading file: %s", err)
	}

	users := make(map[string]struct{})
	err = json.Unmarshal(body, &users)
	if err != nil {
		log.Fatalf("[BACKEND][USERS LOAD] Error unmarshalling file: %s", err)
	}

	Users = UsersType{
		rwmu:  new(sync.RWMutex),
		file:  file,
		users: users,
	}

	defer file.Close()
	<-ctx.Done()

	wg.Done()
}

func GetUsers() map[string]struct{} {
	return Users.users
}

func Login(user string) {
	Users.rwmu.Lock()
	defer Users.rwmu.Unlock()
	Users.users[user] = struct{}{}
	Users.file.WriteString(user + "\n")
}

func Logout(user string) {
	Users.rwmu.Lock()
	defer Users.rwmu.Unlock()
	delete(Users.users, user)

	scanner := bufio.NewScanner(Users.file)
	newFileBuf := make([]byte, 0)

	for scanner.Scan() {
		line := scanner.Text()
		if line != user {
			newFileBuf = append(newFileBuf, line...)
			newFileBuf = append(newFileBuf, '\n')
		}
	}

	Users.file.Truncate(0)
	Users.file.Seek(0, 0)
	Users.file.Write(newFileBuf)
}

// OLD FUNCTIONS
/*
func (users *UsersType) checkLogin(user, password string) bool {
	users.rwmu.RLock()
	defer Users.rwmu.RUnlock()
	if _, ok := Users.users[user]; !ok || Users.users[user] != password {
		return false
	}

	return true
}

func CheckLogin(user, password string) bool {
	return Users.checkLogin(user, password)
}

/*func Login(user, password string) bool {
	return Users.checkLogin(user, password)
}

func (users *UsersType) checkUser(user string) bool {
	users.rwmu.RLock()
	defer Users.rwmu.RUnlock()
	_, ok := Users.users[user]

	return ok
}

func (users *UsersType) register(user, password string) {
	users.rwmu.Lock()
	defer Users.rwmu.Unlock()

	hashed := crypto.Hash(password)

	users.users[user] = hashed
	users.file.WriteString(user + ":" + hashed + "\n")
}

func Register(user, password string) error {
	return remoteServer.UserRegister(user, crypto.Hash(password))
}

func Login(user, password string) error {
	return remoteServer.UserLogin(user, password)
}*/
