package xrt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Database represents the interface for interacting with the database
type Database struct {
	db *gorm.DB
}

// NewDatabase creates a new database or loads an existing one
func NewDatabase(dbPath string) (*Database, error) {
	isNewDB := !fileExists(dbPath)

	err := os.MkdirAll(filepath.Dir(dbPath), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	database := &Database{db: db}

	err = database.migrate()
	if err != nil {
		return nil, fmt.Errorf("failed to perform migrations: %w", err)
	}

	if isNewDB {
		err = database.createAdminUser()
		if err != nil {
			return nil, fmt.Errorf("failed to create admin user: %w", err)
		}
	}

	return database, nil
}

// migrate performs GORM migrations for all data types
func (d *Database) migrate() error {
	return d.db.AutoMigrate(
		&Resource{},
		&TokenData{},
		&Token{},
		&Passport{},
		&User{},
		&Message{},
		&Session{},
		&Conversation{},
		&ModelSettings{},
		&SystemTool{},
		&Agent{},
		&Konsensus{},
		&StatementOfIntent{},
		&TopLevelCategory{},
		&Project{},
	)
}

// createAdminUser creates the initial admin user
func (d *Database) createAdminUser() error {
	adminUser, err := NewUser("admin", "admin", "admin@local", false)
	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	result := d.db.Create(adminUser)
	if result.Error != nil {
		return fmt.Errorf("failed to save admin user to database: %w", result.Error)
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Generic CRUD operations

func (d *Database) create(value interface{}) error {
	result := d.db.Create(value)
	if result.Error != nil {
		return fmt.Errorf("failed to create record: %w", result.Error)
	}
	return nil
}

func (d *Database) read(id string, value interface{}) error {
	result := d.db.First(value, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("record not found")
		}
		return fmt.Errorf("failed to retrieve record: %w", result.Error)
	}
	return nil
}

func (d *Database) update(value interface{}) error {
	result := d.db.Save(value)
	if result.Error != nil {
		return fmt.Errorf("failed to update record: %w", result.Error)
	}
	return nil
}

func (d *Database) delete(value interface{}) error {
	result := d.db.Delete(value)
	if result.Error != nil {
		return fmt.Errorf("failed to delete record: %w", result.Error)
	}
	return nil
}

// Resource CRUD operations

func (d *Database) CreateResource(resource *Resource) error {
	return d.create(resource)
}

func (d *Database) GetResourceByID(id string) (*Resource, error) {
	var resource Resource
	err := d.read(id, &resource)
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (d *Database) UpdateResource(resource *Resource) error {
	return d.update(resource)
}

func (d *Database) DeleteResource(id string) error {
	resource := &Resource{ID: id}
	return d.delete(resource)
}

// TokenData CRUD operations

func (d *Database) CreateTokenData(tokenData *TokenData) error {
	return d.create(tokenData)
}

func (d *Database) GetTokenDataByID(id string) (*TokenData, error) {
	var tokenData TokenData
	err := d.read(id, &tokenData)
	if err != nil {
		return nil, err
	}
	return &tokenData, nil
}

func (d *Database) UpdateTokenData(tokenData *TokenData) error {
	return d.update(tokenData)
}

func (d *Database) DeleteTokenData(id string) error {
	tokenData := &TokenData{ID: id}
	return d.delete(tokenData)
}

// Token CRUD operations

func (d *Database) CreateToken(token *Token) error {
	return d.create(token)
}

func (d *Database) GetTokenByID(id uint) (*Token, error) {
	var token Token
	result := d.db.First(&token, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to retrieve token: %w", result.Error)
	}
	return &token, nil
}

func (d *Database) UpdateToken(token *Token) error {
	return d.update(token)
}

func (d *Database) DeleteToken(id uint) error {
	token := &Token{Model: gorm.Model{ID: id}}
	return d.delete(token)
}

// Passport CRUD operations

func (d *Database) CreatePassport(passport *Passport) error {
	return d.create(passport)
}

func (d *Database) GetPassportByID(id string) (*Passport, error) {
	var passport Passport
	err := d.read(id, &passport)
	if err != nil {
		return nil, err
	}
	return &passport, nil
}

func (d *Database) UpdatePassport(passport *Passport) error {
	return d.update(passport)
}

func (d *Database) DeletePassport(id string) error {
	passport := &Passport{ID: id}
	return d.delete(passport)
}

// User CRUD operations

func (d *Database) CreateUser(user *User) error {
	return d.create(user)
}

func (d *Database) GetUserByID(id string) (*User, error) {
	var user User
	err := d.read(id, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *Database) UpdateUser(user *User) error {
	return d.update(user)
}

func (d *Database) DeleteUser(id string) error {
	user := &User{ID: id}
	return d.delete(user)
}

// Message CRUD operations

func (d *Database) CreateMessage(message *Message) error {
	return d.create(message)
}

func (d *Database) GetMessageByID(id string) (*Message, error) {
	var message Message
	err := d.read(id, &message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (d *Database) UpdateMessage(message *Message) error {
	return d.update(message)
}

func (d *Database) DeleteMessage(id string) error {
	message := &Message{MessageID: id}
	return d.delete(message)
}

// Session CRUD operations

func (d *Database) CreateSession(session *Session) error {
	return d.create(session)
}

func (d *Database) GetSessionByID(id string) (*Session, error) {
	var session Session
	err := d.read(id, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (d *Database) UpdateSession(session *Session) error {
	return d.update(session)
}

func (d *Database) DeleteSession(id string) error {
	session := &Session{ID: id}
	return d.delete(session)
}

// Conversation CRUD operations

func (d *Database) CreateConversation(conversation *Conversation) error {
	return d.create(conversation)
}

func (d *Database) GetConversationByID(id string) (*Conversation, error) {
	var conversation Conversation
	err := d.read(id, &conversation)
	if err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (d *Database) UpdateConversation(conversation *Conversation) error {
	return d.update(conversation)
}

func (d *Database) DeleteConversation(id string) error {
	conversation := &Conversation{ID: id}
	return d.delete(conversation)
}

// ModelSettings CRUD operations

func (d *Database) CreateModelSettings(modelSettings *ModelSettings) error {
	return d.create(modelSettings)
}

func (d *Database) GetModelSettingsByID(id uint) (*ModelSettings, error) {
	var modelSettings ModelSettings
	result := d.db.First(&modelSettings, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("model settings not found")
		}
		return nil, fmt.Errorf("failed to retrieve model settings: %w", result.Error)
	}
	return &modelSettings, nil
}

func (d *Database) UpdateModelSettings(modelSettings *ModelSettings) error {
	return d.update(modelSettings)
}

func (d *Database) DeleteModelSettings(id uint) error {
	modelSettings := &ModelSettings{Model: gorm.Model{ID: id}}
	return d.delete(modelSettings)
}

// SystemTool CRUD operations

func (d *Database) CreateSystemTool(systemTool *SystemTool) error {
	return d.create(systemTool)
}

func (d *Database) GetSystemToolByID(id string) (*SystemTool, error) {
	var systemTool SystemTool
	err := d.read(id, &systemTool)
	if err != nil {
		return nil, err
	}
	return &systemTool, nil
}

func (d *Database) UpdateSystemTool(systemTool *SystemTool) error {
	return d.update(systemTool)
}

func (d *Database) DeleteSystemTool(id string) error {
	systemTool := &SystemTool{ID: id}
	return d.delete(systemTool)
}

// Agent CRUD operations

func (d *Database) CreateAgent(agent *Agent) error {
	return d.create(agent)
}

func (d *Database) GetAgentByID(id string) (*Agent, error) {
	var agent Agent
	err := d.read(id, &agent)
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (d *Database) UpdateAgent(agent *Agent) error {
	return d.update(agent)
}

func (d *Database) DeleteAgent(id string) error {
	agent := &Agent{AgentID: id}
	return d.delete(agent)
}

// Konsensus CRUD operations

func (d *Database) CreateKonsensus(konsensus *Konsensus) error {
	return d.create(konsensus)
}

func (d *Database) GetKonsensusByID(id string) (*Konsensus, error) {
	var konsensus Konsensus
	err := d.read(id, &konsensus)
	if err != nil {
		return nil, err
	}
	return &konsensus, nil
}

func (d *Database) UpdateKonsensus(konsensus *Konsensus) error {
	return d.update(konsensus)
}

func (d *Database) DeleteKonsensus(id string) error {
	konsensus := &Konsensus{ID: id}
	return d.delete(konsensus)
}

// StatementOfIntent CRUD operations

func (d *Database) CreateStatementOfIntent(soi *StatementOfIntent) error {
	return d.create(soi)
}

func (d *Database) GetStatementOfIntentByID(id string) (*StatementOfIntent, error) {
	var soi StatementOfIntent
	err := d.read(id, &soi)
	if err != nil {
		return nil, err
	}
	return &soi, nil
}

func (d *Database) UpdateStatementOfIntent(soi *StatementOfIntent) error {
	return d.update(soi)
}

func (d *Database) DeleteStatementOfIntent(id string) error {
	soi := &StatementOfIntent{ID: id}
	return d.delete(soi)
}

// TopLevelCategory CRUD operations

func (d *Database) CreateTopLevelCategory(category *TopLevelCategory) error {
	return d.create(category)
}

func (d *Database) GetTopLevelCategoryByID(id string) (*TopLevelCategory, error) {
	var category TopLevelCategory
	err := d.read(id, &category)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (d *Database) UpdateTopLevelCategory(category *TopLevelCategory) error {
	return d.update(category)
}

func (d *Database) DeleteTopLevelCategory(id string) error {
	category := &TopLevelCategory{ID: id}
	return d.delete(category)
}

// Project CRUD operations

func (d *Database) CreateProject(project *Project) error {
	return d.create(project)
}

func (d *Database) GetProjectByID(id string) (*Project, error) {
	var project Project
	err := d.read(id, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (d *Database) UpdateProject(project *Project) error {
	return d.update(project)
}

func (d *Database) DeleteProject(id string) error {
	project := &Project{ID: id}
	return d.delete(project)
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}
