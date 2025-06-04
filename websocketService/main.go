package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	socketHandler "github.com/vickypl/chat-app/websocketService/socket/connection"
	connStore "github.com/vickypl/chat-app/websocketService/store/connection"
)

func init() {
	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	messageWriter := &kafka.Writer{
		Addr:     kafka.TCP(os.Getenv("KAFKA_HOST")),
		Topic:    os.Getenv("KAFKA_TOPIC"),
		Balancer: &kafka.LeastBytes{},
	}
	defer messageWriter.Close()

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	connectionStore := connStore.NewInMemoryStore()
	websocketHandler := socketHandler.NewChatWebSocketHandler(connectionStore, &upgrader, messageWriter)

	router := mux.NewRouter()
	router.Handle("/ws/{userid}", authMiddleware(http.HandlerFunc(websocketHandler.HandleConnection)))
	log.Println("WebSocketChat server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func authMiddleware(next http.Handler) http.Handler {
	secretKey := os.Getenv("SECRET_KEY")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		var tokenStr string
		_, err := fmt.Sscanf(authHeader, "Bearer %s", &tokenStr)
		if err != nil {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		username, err := verifyJWT([]byte(secretKey), tokenStr)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Printf("user: %v authenticated successfully..", username)
		next.ServeHTTP(w, r)
	})
}

func verifyJWT(jwtSecret []byte, tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if username, ok := claims["username"].(string); ok {
			return username, nil
		}
	}

	return "", jwt.ErrSignatureInvalid
}
