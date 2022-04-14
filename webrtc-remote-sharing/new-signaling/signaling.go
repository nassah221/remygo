package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/remygo/new-signaling/hub"
	"github.com/remygo/new-signaling/hub/handler"
)

var (
	addr = flag.String("addr", ":8765", "http service address")
)

const apiChanBuffer = 1024

func main() {
	flag.Parse()

	apiChan := make(chan handler.APICall, apiChanBuffer)
	h := hub.New(apiChan)
	go h.StartAPIService(apiChan)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		h.Serve(w, r)
	})

	log.Println("Listening on address:", *addr)
	panic(http.ListenAndServe(*addr, nil))
}
