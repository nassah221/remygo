package hub

import (
	"net/http"

	"github.com/remygo/new-signaling/hub/handler"
)

type Hub struct {
	manager *handler.Manager
}

func New(apiChan chan handler.APICall) *Hub {
	return &Hub{handler.NewManager(apiChan)}
}

func (h *Hub) Serve(w http.ResponseWriter, r *http.Request) {
	h.manager.ServeWs(w, r)
}

func (h *Hub) StartAPIService(apiChan chan handler.APICall) {
	h.manager.LoggingService(apiChan)
}
