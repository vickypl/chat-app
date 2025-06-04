package message

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vickypl/chat-app/persistenceService/model"
	"github.com/vickypl/chat-app/persistenceService/processor"
	"github.com/vickypl/chat-app/persistenceService/store"
)

const (
	retryHeaderKey = "retry-count"
)

type messageProcessor struct {
	maxRetryValue                  int
	messageStore                   store.MessageStore
	reader                         *kafka.Reader
	retryWriter, deadMessageWriter *kafka.Writer
}

func NewMessageProcessor(maxRetryValue int, messageStore store.MessageStore, reader *kafka.Reader, retryWriter, deadMessageWriter *kafka.Writer) processor.MessageProcessor {
	return &messageProcessor{
		maxRetryValue:     maxRetryValue,
		messageStore:      messageStore,
		reader:            reader,
		retryWriter:       retryWriter,
		deadMessageWriter: deadMessageWriter,
	}
}

func (mp *messageProcessor) StartMessageProcessor() {
	for {
		msg, err := mp.reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading message: %v\n", err)
			continue
		}

		var message model.Message
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Printf("Error unmarshaling message: %v\n", err)
			continue
		}

		if err := mp.messageStore.StoreMessage(&message); err != nil {
			log.Printf("DB store failed for message ID %s: %v\n", message.ID, err)
			mp.handleFailedMessage(msg, mp.retryWriter, mp.deadMessageWriter)
			continue
		}

		if err := mp.reader.CommitMessages(context.Background(), msg); err != nil {
			log.Printf("Failed to commit message offset: %v\n", err)
		} else {
			log.Printf("Stored & committed message ID: %s\n", message.ID)
		}
	}
}

func (mp *messageProcessor) handleFailedMessage(msg kafka.Message, producer, dlqWriter *kafka.Writer) {
	retryCount := getRetryCount(msg.Headers)
	if retryCount >= mp.maxRetryValue {
		log.Printf("Max retry reached. Sending to dead-messages. Message ID: %s\n", msg.Key)
		sendToTopic(dlqWriter, msg.Value)
	} else {
		newRetryCount := retryCount + 1
		log.Printf("Retrying message. Attempt #%d\n", newRetryCount)
		msg.Headers = setRetryCount(msg.Headers, newRetryCount)
		sendToTopic(producer, msg.Value, msg.Headers...)
	}
}

func getRetryCount(headers []kafka.Header) int {
	for _, h := range headers {
		if h.Key == retryHeaderKey {
			count, err := strconv.Atoi(string(h.Value))
			if err == nil {
				return count
			}
		}
	}

	return 0
}

func setRetryCount(headers []kafka.Header, count int) []kafka.Header {
	found := false
	for i := range headers {
		if headers[i].Key == retryHeaderKey {
			headers[i].Value = []byte(strconv.Itoa(count))
			found = true
			break
		}
	}

	if !found {
		headers = append(headers, kafka.Header{Key: retryHeaderKey, Value: []byte(strconv.Itoa(count))})
	}

	return headers
}

func sendToTopic(writer *kafka.Writer, value []byte, headers ...kafka.Header) {
	err := writer.WriteMessages(context.Background(), kafka.Message{
		Value:   value,
		Time:    time.Now(),
		Headers: headers,
	})

	if err != nil {
		log.Printf("Failed to publish message: %v\n", err)
	}
}
