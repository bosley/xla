package xrt

import (
	"encoding/json"
	"errors"
)

// StatementOfIntentManager handles operations related to StatementOfIntent
type StatementOfIntentManager struct {
	db *Database
}

// NewStatementOfIntentManager creates a new StatementOfIntentManager
func NewStatementOfIntentManager(db *Database) *StatementOfIntentManager {
	return &StatementOfIntentManager{db: db}
}

// CreateStatementOfIntent creates a new StatementOfIntent
func (soim *StatementOfIntentManager) CreateStatementOfIntent(initialStatement, refinedStatement string, refiningSteps, explicitExpectations, pointsOfConcernListed []string) (*StatementOfIntent, error) {
	if initialStatement == "" || refinedStatement == "" {
		return nil, errors.New("initialStatement and refinedStatement cannot be empty")
	}

	refiningStepsJSON, err := json.Marshal(refiningSteps)
	if err != nil {
		return nil, err
	}

	explicitExpectationsJSON, err := json.Marshal(explicitExpectations)
	if err != nil {
		return nil, err
	}

	pointsOfConcernListedJSON, err := json.Marshal(pointsOfConcernListed)
	if err != nil {
		return nil, err
	}

	soi := &StatementOfIntent{
		InitialStatement:      initialStatement,
		RefinedStatement:      refinedStatement,
		RefiningSteps:         string(refiningStepsJSON),
		ExplicitExpectations:  string(explicitExpectationsJSON),
		PointsOfConcernListed: string(pointsOfConcernListedJSON),
	}

	err = soim.db.CreateStatementOfIntent(soi)
	if err != nil {
		return nil, err
	}

	return soi, nil
}

// GetStatementOfIntent retrieves a StatementOfIntent by ID
func (soim *StatementOfIntentManager) GetStatementOfIntent(id string) (*StatementOfIntent, error) {
	return soim.db.GetStatementOfIntentByID(id)
}

// UpdateStatementOfIntent updates an existing StatementOfIntent
func (soim *StatementOfIntentManager) UpdateStatementOfIntent(id, initialStatement, refinedStatement string, refiningSteps, explicitExpectations, pointsOfConcernListed []string) error {
	soi, err := soim.GetStatementOfIntent(id)
	if err != nil {
		return err
	}

	soi.InitialStatement = initialStatement
	soi.RefinedStatement = refinedStatement

	refiningStepsJSON, err := json.Marshal(refiningSteps)
	if err != nil {
		return err
	}
	soi.RefiningSteps = string(refiningStepsJSON)

	explicitExpectationsJSON, err := json.Marshal(explicitExpectations)
	if err != nil {
		return err
	}
	soi.ExplicitExpectations = string(explicitExpectationsJSON)

	pointsOfConcernListedJSON, err := json.Marshal(pointsOfConcernListed)
	if err != nil {
		return err
	}
	soi.PointsOfConcernListed = string(pointsOfConcernListedJSON)

	return soim.db.UpdateStatementOfIntent(soi)
}

// DeleteStatementOfIntent deletes a StatementOfIntent by ID
func (soim *StatementOfIntentManager) DeleteStatementOfIntent(id string) error {
	return soim.db.DeleteStatementOfIntent(id)
}
