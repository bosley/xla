package xrt

import (
	"encoding/json"
	"errors"
)

// Resource represents files and resources used
type ResourceManager struct {
	db *Database
}

// NewResourceManager creates a new ResourceManager
func NewResourceManager(db *Database) *ResourceManager {
	return &ResourceManager{db: db}
}

// CreateResource creates a new resource
func (rm *ResourceManager) CreateResource(rType byte, value string, meta map[string]interface{}) (*Resource, error) {
	if value == "" {
		return nil, errors.New("resource value cannot be empty")
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}

	resource := &Resource{
		RType: rType,
		Value: value,
		Meta:  string(metaJSON),
	}

	err = rm.db.CreateResource(resource)
	if err != nil {
		return nil, err
	}

	return resource, nil
}

// GetResource retrieves a resource by ID
func (rm *ResourceManager) GetResource(id string) (*Resource, error) {
	return rm.db.GetResourceByID(id)
}

// UpdateResource updates an existing resource
func (rm *ResourceManager) UpdateResource(id string, rType byte, value string, meta map[string]interface{}) error {
	resource, err := rm.GetResource(id)
	if err != nil {
		return err
	}

	resource.RType = rType
	resource.Value = value

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	resource.Meta = string(metaJSON)

	return rm.db.UpdateResource(resource)
}

// DeleteResource deletes a resource by ID
func (rm *ResourceManager) DeleteResource(id string) error {
	return rm.db.DeleteResource(id)
}

// GetResourceMeta retrieves the metadata of a resource as a map
func (rm *ResourceManager) GetResourceMeta(id string) (map[string]interface{}, error) {
	resource, err := rm.GetResource(id)
	if err != nil {
		return nil, err
	}

	var meta map[string]interface{}
	err = json.Unmarshal([]byte(resource.Meta), &meta)
	if err != nil {
		return nil, err
	}

	return meta, nil
}
