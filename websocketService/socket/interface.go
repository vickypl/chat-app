package socket

import "net/http"

type ChatWebSocketHandler interface {
	HandleConnection(w http.ResponseWriter, r *http.Request)
}
