package connection

import (
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vickypl/chat-app/websocketService/store"
)

type inMemoryStore struct {
	mu    *sync.Mutex
	store map[uuid.UUID]*websocket.Conn
}

func NewInMemoryStore() store.ConnectionStore {
	mutex := sync.Mutex{}
	return &inMemoryStore{mu: &mutex, store: make(map[uuid.UUID]*websocket.Conn)}
}

func (ims *inMemoryStore) StoreConnection(key uuid.UUID, value *websocket.Conn) {
	if _, ok := ims.store[key]; ok {
		log.Printf("already connected with userID: %v", key)
		return
	}

	ims.mu.Lock()
	ims.store[key] = value
	ims.mu.Unlock()
	log.Printf("connection stored for userID: %v", key)
}

func (ims *inMemoryStore) GetConnection(key uuid.UUID) *websocket.Conn {
	connection, found := ims.store[key]
	if !found {
		return nil
	}

	return connection
}
