package xrt

import (
	"encoding/json"
	"errors"
)

// SystemToolManager handles operations related to SystemTool
type SystemToolManager struct {
	db *Database
}

// NewSystemToolManager creates a new SystemToolManager
func NewSystemToolManager(db *Database) *SystemToolManager {
	return &SystemToolManager{db: db}
}

// CreateSystemTool creates a new SystemTool
func (stm *SystemToolManager) CreateSystemTool(description, agentPrompt, resourceID string, activations, tags []string) (*SystemTool, error) {
	if description == "" || agentPrompt == "" || resourceID == "" {
		return nil, errors.New("description, agentPrompt, and resourceID cannot be empty")
	}

	activationsJSON, err := json.Marshal(activations)
	if err != nil {
		return nil, err
	}

	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}

	systemTool := &SystemTool{
		Description: description,
		AgentPrompt: agentPrompt,
		ResourceID:  resourceID,
		Activations: string(activationsJSON),
		Tags:        string(tagsJSON),
	}

	err = stm.db.CreateSystemTool(systemTool)
	if err != nil {
		return nil, err
	}

	return systemTool, nil
}

// GetSystemTool retrieves a SystemTool by ID
func (stm *SystemToolManager) GetSystemTool(id string) (*SystemTool, error) {
	return stm.db.GetSystemToolByID(id)
}

// UpdateSystemTool updates an existing SystemTool
func (stm *SystemToolManager) UpdateSystemTool(id, description, agentPrompt, resourceID string, activations, tags []string) error {
	systemTool, err := stm.GetSystemTool(id)
	if err != nil {
		return err
	}

	systemTool.Description = description
	systemTool.AgentPrompt = agentPrompt
	systemTool.ResourceID = resourceID

	activationsJSON, err := json.Marshal(activations)
	if err != nil {
		return err
	}
	systemTool.Activations = string(activationsJSON)

	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return err
	}
	systemTool.Tags = string(tagsJSON)

	return stm.db.UpdateSystemTool(systemTool)
}

// DeleteSystemTool deletes a SystemTool by ID
func (stm *SystemToolManager) DeleteSystemTool(id string) error {
	return stm.db.DeleteSystemTool(id)
}
