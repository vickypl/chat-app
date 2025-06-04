package store

import "github.com/vickypl/chat-app/persistenceService/model"

type MessageStore interface {
	StoreMessage(msg *model.Message) error
}
