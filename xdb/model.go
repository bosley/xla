package xrt

import (
	"errors"
)

// ModelSettingsManager handles operations related to ModelSettings
type ModelSettingsManager struct {
	db *Database
}

// NewModelSettingsManager creates a new ModelSettingsManager
func NewModelSettingsManager(db *Database) *ModelSettingsManager {
	return &ModelSettingsManager{db: db}
}

// CreateModelSettings creates a new ModelSettings
func (msm *ModelSettingsManager) CreateModelSettings(temperature float32, modelNameTag, setupPrompt string, messagePreLoadID string) (*ModelSettings, error) {
	if modelNameTag == "" {
		return nil, errors.New("modelNameTag cannot be empty")
	}

	modelSettings := &ModelSettings{
		Temperature:      temperature,
		ModelNameTag:     modelNameTag,
		SetupPrompt:      setupPrompt,
		MessagePreLoadID: messagePreLoadID,
	}

	err := msm.db.CreateModelSettings(modelSettings)
	if err != nil {
		return nil, err
	}

	return modelSettings, nil
}

// GetModelSettings retrieves a ModelSettings by ID
func (msm *ModelSettingsManager) GetModelSettings(id uint) (*ModelSettings, error) {
	return msm.db.GetModelSettingsByID(id)
}

// UpdateModelSettings updates an existing ModelSettings
func (msm *ModelSettingsManager) UpdateModelSettings(id uint, temperature float32, modelNameTag, setupPrompt string, messagePreLoadID string) error {
	modelSettings, err := msm.GetModelSettings(id)
	if err != nil {
		return err
	}

	modelSettings.Temperature = temperature
	modelSettings.ModelNameTag = modelNameTag
	modelSettings.SetupPrompt = setupPrompt
	modelSettings.MessagePreLoadID = messagePreLoadID

	return msm.db.UpdateModelSettings(modelSettings)
}

// DeleteModelSettings deletes a ModelSettings by ID
func (msm *ModelSettingsManager) DeleteModelSettings(id uint) error {
	return msm.db.DeleteModelSettings(id)
}
