package xrt

import (
	"errors"
)

// ProjectManager handles operations related to Project
type ProjectManager struct {
	db *Database
}

// NewProjectManager creates a new ProjectManager
func NewProjectManager(db *Database) *ProjectManager {
	return &ProjectManager{db: db}
}

// CreateProject creates a new Project
func (pm *ProjectManager) CreateProject(name string, soiID string) (*Project, error) {
	if name == "" {
		return nil, errors.New("project name cannot be empty")
	}

	project := &Project{
		Name:  name,
		SoiID: soiID,
	}

	err := pm.db.CreateProject(project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// GetProject retrieves a Project by ID
func (pm *ProjectManager) GetProject(id string) (*Project, error) {
	return pm.db.GetProjectByID(id)
}

// UpdateProject updates an existing Project
func (pm *ProjectManager) UpdateProject(id string, name string, soiID string) error {
	project, err := pm.GetProject(id)
	if err != nil {
		return err
	}

	project.Name = name
	project.SoiID = soiID

	return pm.db.UpdateProject(project)
}

// DeleteProject deletes a Project by ID
func (pm *ProjectManager) DeleteProject(id string) error {
	return pm.db.DeleteProject(id)
}

// AddAgentToProject adds an Agent to a Project
func (pm *ProjectManager) AddAgentToProject(projectID string, agentID string) error {
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	agent, err := pm.db.GetAgentByID(agentID)
	if err != nil {
		return err
	}

	return pm.db.db.Model(project).Association("Agents").Append(agent)
}

// AddUserToProject adds a non-agent User to a Project
func (pm *ProjectManager) AddUserToProject(projectID string, userID string) error {
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	user, err := pm.db.GetUserByID(userID)
	if err != nil {
		return err
	}

	return pm.db.db.Model(project).Association("NonAgentUsers").Append(user)
}

// AddResourceToProject adds a Resource to a Project
func (pm *ProjectManager) AddResourceToProject(projectID string, resourceID string) error {
	project, err := pm.GetProject(projectID)
	if err != nil {
		return err
	}

	resource, err := pm.db.GetResourceByID(resourceID)
	if err != nil {
		return err
	}

	return pm.db.db.Model(project).Association("Resources").Append(resource)
}
