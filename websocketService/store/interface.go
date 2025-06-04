package store

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type ConnectionStore interface {
	StoreConnection(key uuid.UUID, value *websocket.Conn)
	GetConnection(key uuid.UUID) *websocket.Conn
}
