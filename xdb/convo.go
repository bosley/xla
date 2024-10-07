package xrt

import (
	"errors"
	"time"
)

// ConversationManager handles operations related to Conversation
type ConversationManager struct {
	db *Database
}

// NewConversationManager creates a new ConversationManager
func NewConversationManager(db *Database) *ConversationManager {
	return &ConversationManager{db: db}
}

// CreateConversation creates a new Conversation
func (cm *ConversationManager) CreateConversation(rootMessageID string) (*Conversation, error) {
	if rootMessageID == "" {
		return nil, errors.New("rootMessageID cannot be empty")
	}

	conversation := &Conversation{
		Started:       time.Now(),
		RootMessageID: rootMessageID,
		Active:        true,
	}

	err := cm.db.CreateConversation(conversation)
	if err != nil {
		return nil, err
	}

	return conversation, nil
}

// GetConversation retrieves a Conversation by ID
func (cm *ConversationManager) GetConversation(id string) (*Conversation, error) {
	return cm.db.GetConversationByID(id)
}

// UpdateConversation updates an existing Conversation
func (cm *ConversationManager) UpdateConversation(id string, active bool) error {
	conversation, err := cm.GetConversation(id)
	if err != nil {
		return err
	}

	conversation.Active = active

	return cm.db.UpdateConversation(conversation)
}

// DeleteConversation deletes a Conversation by ID
func (cm *ConversationManager) DeleteConversation(id string) error {
	return cm.db.DeleteConversation(id)
}

// AddParticipantToConversation adds a User to a Conversation
func (cm *ConversationManager) AddParticipantToConversation(conversationID string, userID string) error {
	conversation, err := cm.GetConversation(conversationID)
	if err != nil {
		return err
	}

	user, err := cm.db.GetUserByID(userID)
	if err != nil {
		return err
	}

	return cm.db.db.Model(conversation).Association("Participants").Append(user)
}

// AddMessageToConversation adds a Message to a Conversation
func (cm *ConversationManager) AddMessageToConversation(conversationID string, messageID string) error {
	conversation, err := cm.GetConversation(conversationID)
	if err != nil {
		return err
	}

	message, err := cm.db.GetMessageByID(messageID)
	if err != nil {
		return err
	}

	return cm.db.db.Model(conversation).Association("Messages").Append(message)
}
