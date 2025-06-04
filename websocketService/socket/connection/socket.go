package connection

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"
	"github.com/vickypl/chat-app/websocketService/model"
	"github.com/vickypl/chat-app/websocketService/socket"
	"github.com/vickypl/chat-app/websocketService/store"
)

type chatWebSocketHandler struct {
	store         store.ConnectionStore
	upgrader      *websocket.Upgrader
	messageWriter *kafka.Writer
}

func NewChatWebSocketHandler(store store.ConnectionStore, upgrader *websocket.Upgrader, messageWriter *kafka.Writer) socket.ChatWebSocketHandler {
	return &chatWebSocketHandler{store: store, upgrader: upgrader, messageWriter: messageWriter}
}

func (ch *chatWebSocketHandler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := ch.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	// extracting userID for storing new connection
	vars := mux.Vars(r)
	userIDStr := vars["userid"]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid uuid", http.StatusBadRequest)
		return
	}

	// storing new connection
	ch.store.StoreConnection(userID, conn)

	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		var msg model.Message
		err = json.Unmarshal(msgBytes, &msg)
		if err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}

		err = ch.sendMessageToRecipient(&msg)
		if err != nil {
			log.Println("Failed to send message:", err)
			continue
		}

		kafkaMsg, err := json.Marshal(msg)
		if err != nil {
			log.Println("Failed to marshal message for Kafka:", err)
			continue
		}

		err = ch.messageWriter.WriteMessages(context.Background(),
			kafka.Message{
				Value: kafkaMsg,
			},
		)

		if err != nil {
			log.Println("Failed to write message to Kafka:", err)
			continue
		}

		log.Printf("Message published to Kafka topic 'messages': %+v\n", msg)
	}
}

func (ch *chatWebSocketHandler) sendMessageToRecipient(msg *model.Message) error {
	if msg.ID == uuid.Nil {
		msg.ID = uuid.New()
	}

	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = time.Now()
	}

	recipientConn := ch.store.GetConnection(msg.RecipientID)
	if recipientConn == nil {
		senderConn := ch.store.GetConnection(msg.SenderID)
		serverRepsonse, _ := json.Marshal("Recipient " + msg.RecipientID.String() + " is not online.")
		senderConn.WriteMessage(websocket.TextMessage, serverRepsonse)
		return fmt.Errorf("recipient not found for UUID: %v", msg.RecipientID)
	}

	resp, _ := json.Marshal(msg)
	return recipientConn.WriteMessage(websocket.TextMessage, resp)
}
