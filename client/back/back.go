package back

import (
	"context"
	"log"
	"messengerClient/back/saved"
	"messengerClient/consts"
	"messengerClient/front/handlers"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	errCh := make(chan error)

	startFront(8080, errCh)
	saved.RestoreChats()
	defer saved.SaveChats()

	for {
		select {
		case <-ctx.Done():
			log.Println("[BACKEND][SERVER] Server has been stopped by app exit")
			return

		case err := <-errCh:
			log.Printf("[BACKEND][SERVER] Error crashed the server: %s", err)
			startFront(8080, errCh)
			saved.RestoreChats()
		}
	}
}

func startFront(port int, errCh chan error) {
	log.Printf("[BACKEND][SERVER] Starting")

	ipStarter := ""
	if runtime.GOOS == "linux" {
		ipStarter = consts.LocalIP
		err := exec.Command("xdg-open", "http://"+ipStarter+strconv.Itoa(port)).Start()
		if err != nil {
			log.Printf("[BACKEND][SERVER] Error opening browser: %s", err)
		}
	} else {
		ipStarter = consts.LocalHost
		err := exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://"+ipStarter+strconv.Itoa(port)).Start()
		if err != nil {
			log.Printf("[BACKEND][SERVER] Error opening browser: %s", err)
		}
	}

	server := &http.Server{
		Addr: ipStarter + strconv.Itoa(port),
	}
	serverIsRunning := make(chan bool)
	go func(serverIsRunning chan bool) {
		serverIsRunning <- true
		err := server.ListenAndServe()
		if err != nil {
			errCh <- err
		}
	}(serverIsRunning)
	<-serverIsRunning
	close(serverIsRunning)
	log.Printf("[BACKEND][SERVER] Listening on port %d", port)

	handlers.Init()

	//return server
}
