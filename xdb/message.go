package xrt

import (
	"encoding/json"
	"errors"
	"time"
)

// MessageManager handles operations related to Message
type MessageManager struct {
	db *Database
}

// NewMessageManager creates a new MessageManager
func NewMessageManager(db *Database) *MessageManager {
	return &MessageManager{db: db}
}

// CreateMessage creates a new Message
func (mm *MessageManager) CreateMessage(topicID, idTo, idFrom, contents string, meta map[string]interface{}, previousID string) (*Message, error) {
	if topicID == "" || idTo == "" || idFrom == "" || contents == "" {
		return nil, errors.New("topicID, idTo, idFrom, and contents cannot be empty")
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}

	message := &Message{
		TopicID:    topicID,
		IDTo:       idTo,
		IDFrom:     idFrom,
		Sent:       time.Now(),
		Contents:   contents,
		Meta:       string(metaJSON),
		PreviousID: previousID,
	}

	err = mm.db.CreateMessage(message)
	if err != nil {
		return nil, err
	}

	return message, nil
}

// GetMessage retrieves a Message by ID
func (mm *MessageManager) GetMessage(id string) (*Message, error) {
	return mm.db.GetMessageByID(id)
}

// UpdateMessage updates an existing Message
func (mm *MessageManager) UpdateMessage(id, contents string, meta map[string]interface{}) error {
	message, err := mm.GetMessage(id)
	if err != nil {
		return err
	}

	message.Contents = contents

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	message.Meta = string(metaJSON)

	return mm.db.UpdateMessage(message)
}

// DeleteMessage deletes a Message by ID
func (mm *MessageManager) DeleteMessage(id string) error {
	return mm.db.DeleteMessage(id)
}

// SetMessageReceived sets the received time for a message
func (mm *MessageManager) SetMessageReceived(id string) error {
	message, err := mm.GetMessage(id)
	if err != nil {
		return err
	}

	message.Received = time.Now()

	return mm.db.UpdateMessage(message)
}
