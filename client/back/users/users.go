package users

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"messengerClient/back/saved"
	"messengerClient/consts"
	"os"
	"sync"
)

type UsersType struct {
	rwmu  *sync.RWMutex
	file  *os.File
	users map[string]struct{}
}

var Users UsersType

func RefreshUsers(ctx context.Context, wg *sync.WaitGroup) {
	file, err := os.OpenFile("back/users.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("[BACKEND][USERS LOAD] Error opening file: %s", err)
	}

	file.Truncate(0)
	file.Seek(0, 0)
	file.WriteString("{}")

	Users = UsersType{
		rwmu:  new(sync.RWMutex),
		file:  file,
		users: make(map[string]struct{}),
	}

	defer file.Close()
	<-ctx.Done()

	Users.rwmu.Lock()
	defer Users.rwmu.Unlock()
	saveUsers()

	wg.Done()
}

func saveUsers() {
	body, err := json.Marshal(Users.users)
	log.Printf("[BACKEND][USERS SAVE] Users: %s", string(body))
	if err != nil {
		log.Printf("[BACKEND][USERS SAVE] Error encoding file: %s", err)
		return
	}

	file, err := os.OpenFile("back/users.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("[BACKEND][USERS SAVE] Error opening file: %s", err)
	}
	defer file.Close()

	file.Truncate(0)
	file.Seek(0, 0)
	file.Write(body)
}

func GetUsers() map[string]struct{} {
	return Users.users
}

func Login(user, password string) {
	Users.rwmu.Lock()
	Users.rwmu.Unlock()
	Users.users[user] = struct{}{}

	consts.EventListeners.Mu.Lock()
	consts.EventListeners.Events[user] = make(map[string]map[string]chan int)
	consts.EventListeners.Mu.Unlock()

	consts.Recievers.Mu.Lock()
	consts.Recievers.Events[user] = make(map[string]chan []byte)
	consts.Recievers.Mu.Unlock()

	saved.CheckChats(user, password)

	saveUsers()
}

func Logout(user string) {
	Users.rwmu.Lock()
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

	saved.ClearChats(user)
	saveUsers()
	Users.rwmu.Unlock()

	consts.EventListeners.Mu.Lock()
	delete(consts.EventListeners.Events, user)
	consts.EventListeners.Mu.Unlock()

	consts.Recievers.Mu.Lock()
	delete(consts.Recievers.Events, user)
	consts.Recievers.Mu.Unlock()
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
