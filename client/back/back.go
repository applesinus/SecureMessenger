package back

import (
	"context"
	"fmt"
	"log"
	"messengerClient/front/handlers"
	"net/http"
	"runtime"
	"strconv"
	"sync"
)

const (
	localHost = "localhost:"
	ip        = "127.0.0.1:"
)

func Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	errCh := make(chan error)

	startFront(8080, errCh)

	for {
		select {
		case <-ctx.Done():
			log.Println("[BACKEND][SERVER] Server has been stopped by app exit")
			return

		case err := <-errCh:
			log.Printf("[BACKEND][SERVER] Error crashed the server: %s", err)
			startFront(8080, errCh)
		}
	}
}

func startFront(port int, errCh chan error) {
	log.Println("Starting HTTP Server")
	log.Printf("Listening on port %d", port)

	ipStarter := ""
	if runtime.GOOS == "linux" {
		ipStarter = ip
	} else {
		ipStarter = localHost
	}

	server := &http.Server{
		Addr: ipStarter + strconv.Itoa(port),
	}
	serverIsRunning := make(chan bool)
	go func(serverIsRunning chan bool) {
		fmt.Println("\nSERVER IS RUNNING!")
		serverIsRunning <- true
		err := server.ListenAndServe()
		if err != nil {
			errCh <- err
		}
	}(serverIsRunning)
	<-serverIsRunning
	close(serverIsRunning)

	handlers.Init()

	//return server
}
