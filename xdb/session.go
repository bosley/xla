package xrt

import (
	"errors"
)

// SessionManager handles operations related to Session
type SessionManager struct {
	db *Database
}

// NewSessionManager creates a new SessionManager
func NewSessionManager(db *Database) *SessionManager {
	return &SessionManager{db: db}
}

// CreateSession creates a new Session
func (sm *SessionManager) CreateSession(userID, conversationID string) (*Session, error) {
	if userID == "" || conversationID == "" {
		return nil, errors.New("userID and conversationID cannot be empty")
	}

	session := &Session{
		UserID:         userID,
		ConversationID: conversationID,
	}

	err := sm.db.CreateSession(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession retrieves a Session by ID
func (sm *SessionManager) GetSession(id string) (*Session, error) {
	return sm.db.GetSessionByID(id)
}

// UpdateSession updates an existing Session
func (sm *SessionManager) UpdateSession(id, userID, conversationID string) error {
	session, err := sm.GetSession(id)
	if err != nil {
		return err
	}

	session.UserID = userID
	session.ConversationID = conversationID

	return sm.db.UpdateSession(session)
}

// DeleteSession deletes a Session by ID
func (sm *SessionManager) DeleteSession(id string) error {
	return sm.db.DeleteSession(id)
}

// AddResourceToSession adds a Resource to a Session
func (sm *SessionManager) AddResourceToSession(sessionID string, resourceID string) error {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return err
	}

	resource, err := sm.db.GetResourceByID(resourceID)
	if err != nil {
		return err
	}

	return sm.db.db.Model(session).Association("LocalResources").Append(resource)
}
