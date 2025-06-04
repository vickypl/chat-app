package message

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/vickypl/chat-app/persistenceService/model"
	"github.com/vickypl/chat-app/persistenceService/store"
)

const (
	storeQuery = `INSERT INTO messages (id, sender_id, recipient_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`
)

type messageStore struct {
	pgdb *pgxpool.Pool
}

func NewMessageStore(pgdb *pgxpool.Pool) store.MessageStore {
	return &messageStore{pgdb: pgdb}
}

func (ms *messageStore) StoreMessage(msg *model.Message) error {
	_, err := ms.pgdb.Exec(context.Background(), storeQuery, msg.ID, msg.SenderID, msg.RecipientID, msg.Content, msg.CreatedAt)
	return err
}
