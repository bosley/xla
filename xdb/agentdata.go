package xrt

import (
	"errors"
)

// AgentManager handles operations related to Agent
type AgentManager struct {
	db *Database
}

// NewAgentManager creates a new AgentManager
func NewAgentManager(db *Database) *AgentManager {
	return &AgentManager{db: db}
}

// CreateAgent creates a new Agent
func (am *AgentManager) CreateAgent(userID string, settingsID uint) (*Agent, error) {
	if userID == "" {
		return nil, errors.New("user ID cannot be empty")
	}

	agent := &Agent{
		UserID:     userID,
		SettingsID: settingsID,
	}

	err := am.db.CreateAgent(agent)
	if err != nil {
		return nil, err
	}

	return agent, nil
}

// GetAgent retrieves an Agent by ID
func (am *AgentManager) GetAgent(id string) (*Agent, error) {
	return am.db.GetAgentByID(id)
}

// UpdateAgent updates an existing Agent
func (am *AgentManager) UpdateAgent(id string, userID string, settingsID uint) error {
	agent, err := am.GetAgent(id)
	if err != nil {
		return err
	}

	agent.UserID = userID
	agent.SettingsID = settingsID

	return am.db.UpdateAgent(agent)
}

// DeleteAgent deletes an Agent by ID
func (am *AgentManager) DeleteAgent(id string) error {
	return am.db.DeleteAgent(id)
}

// AddToolToAgent adds a SystemTool to an Agent
func (am *AgentManager) AddToolToAgent(agentID string, toolID string) error {
	agent, err := am.GetAgent(agentID)
	if err != nil {
		return err
	}

	tool, err := am.db.GetSystemToolByID(toolID)
	if err != nil {
		return err
	}

	return am.db.db.Model(agent).Association("ToolsAvailable").Append(tool)
}
