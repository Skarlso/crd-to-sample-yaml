package pkg

import (
	"fmt"
	"os"
	"strings"
)

// ConditionEnhancer handles enhancing CRDs with conditions parsed from API folders.
type ConditionEnhancer struct {
	apiFolder  string
	parser     *ConditionParser
	conditions map[string][]ConditionInfo
}

// NewConditionEnhancer creates a new condition enhancer.
func NewConditionEnhancer(apiFolder string) *ConditionEnhancer {
	return &ConditionEnhancer{
		apiFolder:  apiFolder,
		parser:     NewConditionParser(),
		conditions: make(map[string][]ConditionInfo),
	}
}

// LoadConditions parses conditions from the API folder.
func (e *ConditionEnhancer) LoadConditions() error {
	if e.apiFolder == "" {
		return nil // No API folder specified, nothing to do
	}

	if _, err := os.Stat(e.apiFolder); os.IsNotExist(err) {
		return fmt.Errorf("API folder does not exist: %s", e.apiFolder)
	}

	err := e.parser.ParseGoFiles(e.apiFolder)
	if err != nil {
		return fmt.Errorf("failed to parse conditions from API folder: %w", err)
	}

	e.conditions = e.parser.GetConditions()

	return nil
}

// EnhanceSchemas adds conditions to schemas where CRD kind matches.
func (e *ConditionEnhancer) EnhanceSchemas(schemas []*SchemaType) []*SchemaType {
	if len(e.conditions) == 0 {
		return schemas // No conditions to add
	}

	for _, schema := range schemas {
		if conditions, found := e.findConditionsForKind(schema.Kind); found {
			schema.Conditions = conditions
		}
	}

	return schemas
}

// findConditionsForKind finds conditions that match the given CRD kind.
func (e *ConditionEnhancer) findConditionsForKind(kind string) ([]ConditionInfo, bool) {
	if conditions, exists := e.conditions[kind]; exists {
		return conditions, true
	}

	for crdName, conditions := range e.conditions {
		if e.namesMatch(kind, crdName) {
			return conditions, true
		}
	}

	return nil, false
}

// namesMatch performs matching between CRD kind and condition CRD name.
func (e *ConditionEnhancer) namesMatch(kind, crdName string) bool {
	kind = strings.ToLower(kind)
	crdName = strings.ToLower(crdName)

	return kind == crdName
}
