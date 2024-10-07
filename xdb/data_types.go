package xrt

import (
	"time"

	"gorm.io/gorm"
)

const (
	ResourceTypeWeb byte = iota
	ResourceTypeFile
	ResourceTypeMessageChain
)

// Resource represents files and resources used
type Resource struct {
	gorm.Model
	ID    string `gorm:"primaryKey;type:varchar(36)"` // Unique Resource ID
	RType byte   `gorm:"type:integer"`                // (ResourceType<TYPE> from const block above)
	Value string `gorm:"type:text"`                   // The value of the resource (filepath, url, connection string, message id referencing Message struct, etc)
	Meta  string `gorm:"type:text"`                   // Extension type, url type, stored as JSON
}

// TokenData represents authorization tokens users have to perform actions (unique to user)
type TokenData struct {
	gorm.Model
	ID       string    `gorm:"primaryKey;type:varchar(36)"`
	UserID   string    `gorm:"type:varchar(36);index"` // Store this, not *User, as it will be part of the encoded object
	Granted  time.Time `gorm:"type:datetime"`
	Duration int64     `gorm:"type:integer"` // Duration in nanoseconds
}

// Token represents the token value once "on the wire"
type Token struct {
	gorm.Model
	Sent      time.Time `gorm:"type:datetime"` // Must be within the last minute (server restriction)
	Data      TokenData `gorm:"embedded"`      // Value of actual user token info
	Nonce     uint64    `gorm:"type:integer"`  // Random value sent with
	Signature []byte    `gorm:"type:blob"`     // Signature of the token data for verification
}

// Passport represents a user's passport
type Passport struct {
	gorm.Model
	ID      string `gorm:"primaryKey;type:varchar(36)"` // Unique passport ID
	OwnerID string `gorm:"type:varchar(36);index"`      // The owner of the passport (User ID)
	Owner   User   `gorm:"foreignKey:OwnerID"`
	// TODO: Essentially make a better "badger" badge with what we've learned from Badger
}

// User represents a user in the system
type User struct {
	gorm.Model
	ID           string      `gorm:"primaryKey;type:varchar(36)"` // Unique UUIDv4 to mark user
	Name         string      `gorm:"type:varchar(255);unique"`    // Their display name (no spaces, must be unique)
	PasswordHash []byte      `gorm:"type:blob"`                   // bcrypt hash of user password
	Email        string      `gorm:"type:varchar(255);unique"`    // Must be valid email in form, but we won't validate for now
	IsAgent      bool        `gorm:"type:boolean"`                // True if it's an agent
	Tokens       []TokenData `gorm:"foreignKey:UserID"`           // The data values of the tokens the user created for request auth
	Sessions     []Session   `gorm:"foreignKey:UserID"`           // IDs of all sessions user has going
	Passports    []Passport  `gorm:"foreignKey:OwnerID"`          // User's passports
}

// Message represents messages that can contain message chains from multiple agents interacting
type Message struct {
	gorm.Model
	MessageID  string    `gorm:"primaryKey;type:varchar(36)"` // The unique message ID
	TopicID    string    `gorm:"type:varchar(36);index"`      // Unique ID to denote what the message chain is in regards to (which project)
	IDTo       string    `gorm:"type:varchar(36);index"`      // User ID (Agents are users too)
	IDFrom     string    `gorm:"type:varchar(36);index"`      // User ID (Agents are users too)
	Sent       time.Time `gorm:"type:datetime"`               // Time message left sender
	Received   time.Time `gorm:"type:datetime"`               // Time message received by receiver
	Contents   string    `gorm:"type:text"`                   // Contents of message
	Meta       string    `gorm:"type:text"`                   // Values like "role", "sentiment", etc., stored as JSON
	PreviousID string    `gorm:"type:varchar(36)"`            // The previous message ID
	Previous   *Message  `gorm:"foreignKey:PreviousID"`       // Self-referential relationship to previous message
	NextID     string    `gorm:"type:varchar(36)"`            // The next message ID
	Next       *Message  `gorm:"foreignKey:NextID"`           // Self-referential relationship to next message
}

// Session represents any interface to the system on behalf of "any" user (agents too) that is active
type Session struct {
	gorm.Model
	ID             string       `gorm:"primaryKey;type:varchar(36)"` // UUIDv4
	UserID         string       `gorm:"type:varchar(36);index"`      // References User struct ID
	User           User         `gorm:"foreignKey:UserID"`
	ConversationID string       `gorm:"type:varchar(36)"` // The conversation occurring in the session
	Conversation   Conversation `gorm:"foreignKey:ConversationID"`
	LocalResources []Resource   `gorm:"many2many:session_resources"`
}

// Conversation represents a chat conversation
type Conversation struct {
	gorm.Model
	ID            string    `gorm:"primaryKey;type:varchar(36)"` // Unique ID
	Started       time.Time `gorm:"type:datetime"`               // When chat started
	RootMessageID string    `gorm:"type:varchar(36)"`            // First message of message history
	RootMessage   *Message  `gorm:"foreignKey:RootMessageID"`
	Active        bool      `gorm:"type:boolean"` // If the conversation has not yet been "finished"
	Participants  []User    `gorm:"many2many:conversation_participants"`
	Messages      []Message `gorm:"foreignKey:TopicID"` // Messages in this conversation
}

// ModelSettings represents settings for AI models
type ModelSettings struct {
	gorm.Model
	Temperature      float32  `gorm:"type:real"`
	ModelNameTag     string   `gorm:"type:varchar(255)"`
	SetupPrompt      string   `gorm:"type:text"`
	MessagePreLoadID string   `gorm:"type:varchar(36)"` // ID of the pre-loaded message
	MessagePreLoad   *Message `gorm:"foreignKey:MessagePreLoadID"`
}

// SystemTool represents a tool specially crafted for an AI agent to use
type SystemTool struct {
	gorm.Model
	ID          string   `gorm:"primaryKey;type:varchar(36)"`
	Description string   `gorm:"type:text"`
	AgentPrompt string   `gorm:"type:text"`
	ResourceID  string   `gorm:"type:varchar(36)"`
	Activations string   `gorm:"type:text"` // Stored as JSON array
	Tags        string   `gorm:"type:text"` // Stored as JSON array
	Resource    Resource `gorm:"foreignKey:ResourceID"`
}

// Agent represents the actual agent information
type Agent struct {
	gorm.Model
	AgentID        string        `gorm:"primaryKey;type:varchar(36)"`
	UserID         string        `gorm:"type:varchar(36);index"`
	User           User          `gorm:"foreignKey:UserID"`
	SettingsID     uint          `gorm:"type:integer"`
	Settings       ModelSettings `gorm:"foreignKey:SettingsID"`
	ToolsAvailable []SystemTool  `gorm:"many2many:agent_tools"`
}

// Konsensus represents algorithms for gaining consensus in a multi-agent conversation
type Konsensus struct {
	gorm.Model
	ID               string   `gorm:"primaryKey;type:varchar(36)"`
	Name             string   `gorm:"type:varchar(255)"`
	InstructionSetID string   `gorm:"type:varchar(36)"` // The resource identifier for the script
	InstructionSet   Resource `gorm:"foreignKey:InstructionSetID"`
}

// StatementOfIntent represents a refined task statement
type StatementOfIntent struct {
	gorm.Model
	ID                    string `gorm:"primaryKey;type:varchar(36)"`
	InitialStatement      string `gorm:"type:text"`
	RefinedStatement      string `gorm:"type:text"`
	RefiningSteps         string `gorm:"type:text"` // Stored as JSON array
	ExplicitExpectations  string `gorm:"type:text"` // Stored as JSON array
	PointsOfConcernListed string `gorm:"type:text"` // Stored as JSON array
}

// TopLevelCategory represents a top-level category for tasks
type TopLevelCategory struct {
	gorm.Model
	ID                       string   `gorm:"primaryKey;type:varchar(36)"`
	Name                     string   `gorm:"type:varchar(255);unique"`
	Description              string   `gorm:"type:text"`
	RefinementInstructionsID string   `gorm:"type:varchar(36)"` // ID of the script resource
	RefinementInstructions   Resource `gorm:"foreignKey:RefinementInstructionsID"`
}

// Project represents a user project
type Project struct {
	gorm.Model
	ID            string            `gorm:"primaryKey;type:varchar(36)"`
	Name          string            `gorm:"type:varchar(255)"`
	SoiID         string            `gorm:"type:varchar(36)"`
	Soi           StatementOfIntent `gorm:"foreignKey:SoiID"`
	Agents        []Agent           `gorm:"many2many:project_agents"`
	NonAgentUsers []User            `gorm:"many2many:project_users"`
	Resources     []Resource        `gorm:"many2many:project_resources"`
	Sessions      []Session         `gorm:"many2many:project_sessions"`
}
