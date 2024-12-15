package back

import (
	"context"
	"log"
	"messengerClient/back/saved"
	"messengerClient/front/handlers"
	"net/http"
	"strconv"
	"sync"
)

func Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	errCh := make(chan error)

	startClient(8080, errCh)
	saved.RestoreChats()
	defer saved.SaveChats()

	for {
		select {
		case <-ctx.Done():
			log.Println("[BACKEND][SERVER] Server has been stopped by app exit")
			return

		case err := <-errCh:
			log.Printf("[BACKEND][SERVER] Error crashed the server: %s", err)
			startClient(8080, errCh)
			saved.RestoreChats()
		}
	}
}

func startClient(port int, errCh chan error) {
	log.Printf("[BACKEND][SERVER] Starting")

	/*ipStarter := ""
	if runtime.GOOS == "linux" {
		ipStarter = consts.LocalIP
	} else {
		ipStarter = consts.LocalHost
	}*/

	server := &http.Server{
		Addr: /*ipStarter + */ ":" + strconv.Itoa(port),
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
	log.Printf("[BACKEND][SERVER] Listening on %s", server.Addr)

	handlers.Init()
}
