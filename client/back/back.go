package back

import (
	"context"
	"log"
	"messengerClient/back/saved"
	"messengerClient/consts"
	"messengerClient/front/handlers"
	"net/http"
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
	} else {
		ipStarter = consts.LocalHost
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
