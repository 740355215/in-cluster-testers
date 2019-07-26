package main

import (
	"cloud.tencent.com/tke/in-cluster-tester/sidecar/handlers"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func main() {
	//var seconds int
	//flag.IntVar(&seconds, "t", 10, "the seconds sidecar will sleep")
	//flag.Parse()

	r := mux.NewRouter()
	var s = r.Path("/message").Subrouter()
	s.Methods("GET").HandlerFunc(handlers.ReadMessageHandler)
	s.Methods("POST").HandlerFunc(handlers.WriteMessageHandler)

	var delayCloseRouter = r.Path("/delay").Subrouter()
	delayCloseRouter.Methods("POST").HandlerFunc(handlers.WriteDelayTimeHandler)
	svr := http.Server{
		Addr:    ":8099",
		Handler: r,
	}

	// close sidecar server with delay time
	go func() {
		delaySeconds, open := <-handlers.DelayTimeChan
		if !open {
			fmt.Println("DelayTimeChan is closed.")
			return
		}

		time.Sleep(time.Duration(delaySeconds) * time.Second)

		if err := svr.Close(); err != nil {
			fmt.Printf("close sidecar server failed. %v\n", err)
		}
	}()

	if err := svr.ListenAndServe(); err != nil {
		if err == http.ErrServerClosed {
			fmt.Printf("sidecar server closed with delaying time. %v\n", err)
		} else {
			fmt.Printf("sidecar server closed unexpected: %v\n", err)
		}
	}

	return
}
