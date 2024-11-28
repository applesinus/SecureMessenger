package users

import (
	"bufio"
	"context"
	"log"
	"messengerClient/back/crypto"
	"net/http"
	"os"
	"strings"
	"sync"
)

type UsersType struct {
	rwmu  *sync.RWMutex
	file  *os.File
	users map[string]string
}

var Users UsersType

func LoadUsers(ctx context.Context, wg *sync.WaitGroup) {
	file, err := os.OpenFile("back/users.txt", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	users := make(map[string]string)

	for scanner.Scan() {
		words := strings.Split(scanner.Text(), ":")
		if len(words) == 2 {
			users[words[0]] = words[1]
		}
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

func Login(user, password string) bool {
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

func Register(w http.ResponseWriter, user, password string) string {
	if Users.checkUser(user) {
		return "user exists"
	}

	Users.register(user, password)
	return "ok"
}
