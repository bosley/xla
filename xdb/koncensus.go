package xrt

import (
	"errors"
)

// KonsensusManager handles operations related to Konsensus
type KonsensusManager struct {
	db *Database
}

// NewKonsensusManager creates a new KonsensusManager
func NewKonsensusManager(db *Database) *KonsensusManager {
	return &KonsensusManager{db: db}
}

// CreateKonsensus creates a new Konsensus
func (km *KonsensusManager) CreateKonsensus(name string, instructionSetID string) (*Konsensus, error) {
	if name == "" {
		return nil, errors.New("konsensus name cannot be empty")
	}

	konsensus := &Konsensus{
		Name:             name,
		InstructionSetID: instructionSetID,
	}

	err := km.db.CreateKonsensus(konsensus)
	if err != nil {
		return nil, err
	}

	return konsensus, nil
}

// GetKonsensus retrieves a Konsensus by ID
func (km *KonsensusManager) GetKonsensus(id string) (*Konsensus, error) {
	return km.db.GetKonsensusByID(id)
}

// UpdateKonsensus updates an existing Konsensus
func (km *KonsensusManager) UpdateKonsensus(id string, name string, instructionSetID string) error {
	konsensus, err := km.GetKonsensus(id)
	if err != nil {
		return err
	}

	konsensus.Name = name
	konsensus.InstructionSetID = instructionSetID

	return km.db.UpdateKonsensus(konsensus)
}

// DeleteKonsensus deletes a Konsensus by ID
func (km *KonsensusManager) DeleteKonsensus(id string) error {
	return km.db.DeleteKonsensus(id)
}

// ListKonsensus retrieves all Konsensus entries
func (km *KonsensusManager) ListKonsensus() ([]Konsensus, error) {
	var konsensuses []Konsensus
	result := km.db.db.Find(&konsensuses)
	if result.Error != nil {
		return nil, result.Error
	}
	return konsensuses, nil
}

// GetKonsensusByName retrieves a Konsensus by its name
func (km *KonsensusManager) GetKonsensusByName(name string) (*Konsensus, error) {
	var konsensus Konsensus
	result := km.db.db.Where("name = ?", name).First(&konsensus)
	if result.Error != nil {
		return nil, result.Error
	}
	return &konsensus, nil
}
