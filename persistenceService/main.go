package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	msgProcessor "github.com/vickypl/chat-app/persistenceService/processor/message"
	msgStore "github.com/vickypl/chat-app/persistenceService/store/message"
)

func init() {
	err := godotenv.Load("configs/.env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
}

func main() {
	pgSQl, err := getDBConnection()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pgSQl.Close()

	if err := createTable(pgSQl); err != nil {
		log.Fatalf("Failed to create table: %v\n", err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{os.Getenv("KAFKA_HOST")},
		Topic:    os.Getenv("KAFKA_TOPIC"),
		GroupID:  os.Getenv("KAFKA_CONSUMER_GROUP_ID"),
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	// For retrying we are currently publishing it to same topic that is messages
	retryWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{os.Getenv("KAFKA_HOST")},
		Topic:   os.Getenv("KAFKA_TOPIC"),
	})
	defer retryWriter.Close()

	deadMessageWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{os.Getenv("KAFKA_HOST")},
		Topic:   os.Getenv("KAFKA_DEAD_TOPIC"),
	})
	defer deadMessageWriter.Close()

	maxRetryValue := os.Getenv("MAX_RETRY_VALUE")
	maxRetryValueInt, err := strconv.Atoi(maxRetryValue)
	if err != nil {
		log.Fatalf("Invalid max retry value: %v", maxRetryValue)
	}

	// msg store for persistent storage
	messageStore := msgStore.NewMessageStore(pgSQl)
	messageProcessor := msgProcessor.NewMessageProcessor(maxRetryValueInt, messageStore, reader, retryWriter, deadMessageWriter)

	log.Print("Started message processor...\n")
	messageProcessor.StartMessageProcessor()
}

func getDBConnection() (*pgxpool.Pool, error) {
	dbDialect := os.Getenv("DB_DAILECT")
	dbhost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")
	sslMode := os.Getenv("DB_SSL_MODE")

	connStr := dbDialect + "://" + dbUser + ":" + dbPassword + "@" + dbhost + ":" + dbPort + "/" + dbName + "?sslmode=" + sslMode
	return pgxpool.Connect(context.Background(), connStr)
}

func createTable(pool *pgxpool.Pool) error {
	query := `CREATE TABLE IF NOT EXISTS messages (
		id UUID PRIMARY KEY,
		sender_id UUID NOT NULL,
		recipient_id UUID NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL
	)`

	_, err := pool.Exec(context.Background(), query)

	return err
}
