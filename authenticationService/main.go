// auth_service/main.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *authenticator) generateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour).Unix(), // token will be valid for an hour
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secretKey)
}

type authenticator struct {
	secretKey []byte
	userDB    map[string]string // for testing purpose
}

func NewAuthenticator(userDB map[string]string, secretKey []byte) authenticator {
	return authenticator{userDB: userDB, secretKey: secretKey}
}

func (a *authenticator) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userPwdInDB, ok := a.userDB[user.Username]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if userPwdInDB != user.Password {
		http.Error(w, "Incorrect user password", http.StatusUnauthorized)
		return
	}

	token, err := a.generateJWT(user.Username)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func main() {
	// some hard coded dummy user stored for testing purpose
	userDB := make(map[string]string)
	userDB["user1"] = "pwd1"
	userDB["user2"] = "pwd2"
	userDB["user3"] = "pwd3"
	userDB["user4"] = "pwd4"
	userDB["user5"] = "pwd5"

	authHandler := NewAuthenticator(userDB, []byte(os.Getenv("SECRET_KEY")))

	router := mux.NewRouter()
	router.HandleFunc("/login", authHandler.loginHandler)
	log.Println("Authentication server started on :8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
