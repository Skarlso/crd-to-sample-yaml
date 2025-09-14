package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConditionEnhancer_LoadConditions(t *testing.T) {
	// Create a temporary directory with test Go files
	tempDir := t.TempDir()

	testFile := `package types

// AppConditionType defines condition types
type AppConditionType string

const (
	// +cty:condition:for:App
	// Frontend is ready for requests
	FrontendReady AppConditionType = "FrontendReady"
)

// AppReasonType defines reasons
type AppReasonType string

const (
	// +cty:reason:for:App/FrontendReady
	// Service is healthy
	Healthy AppReasonType = "Healthy"
)
`

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(testFile), 0644))

	enhancer := NewConditionEnhancer(tempDir)
	err := enhancer.LoadConditions()
	require.NoError(t, err)

	assert.Len(t, enhancer.conditions, 1)
	assert.Contains(t, enhancer.conditions, "App")
	assert.Len(t, enhancer.conditions["App"], 1)
	assert.Equal(t, "FrontendReady", enhancer.conditions["App"][0].Type)
	assert.Len(t, enhancer.conditions["App"][0].Reasons, 1)
}

func TestConditionEnhancer_LoadConditions_EmptyFolder(t *testing.T) {
	enhancer := NewConditionEnhancer("")
	err := enhancer.LoadConditions()
	require.NoError(t, err)
	assert.Empty(t, enhancer.conditions)
}

func TestConditionEnhancer_LoadConditions_NonExistentFolder(t *testing.T) {
	enhancer := NewConditionEnhancer("/non/existent/path")
	err := enhancer.LoadConditions()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API folder does not exist")
}

func TestConditionEnhancer_EnhanceSchemas(t *testing.T) {
	tempDir := t.TempDir()

	testFile := `package types

const (
	// +cty:condition:for:MyApp
	// Service is ready
	ServiceReady = "ServiceReady"
)

const (
	// +cty:reason:for:MyApp/ServiceReady
	// All checks passed
	AllHealthy = "Healthy"
)
`

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(testFile), 0644))

	enhancer := NewConditionEnhancer(tempDir)
	err := enhancer.LoadConditions()
	require.NoError(t, err)

	schemas := []*SchemaType{
		{Kind: "MyApp", Group: "example.com"},
		{Kind: "OtherApp", Group: "example.com"},
	}
	enhanced := enhancer.EnhanceSchemas(schemas)

	assert.Len(t, enhanced[0].Conditions, 1)
	assert.Equal(t, "ServiceReady", enhanced[0].Conditions[0].Type)
	assert.Len(t, enhanced[0].Conditions[0].Reasons, 1)
	assert.Empty(t, enhanced[1].Conditions)
}

func TestConditionEnhancer_NamesMatch(t *testing.T) {
	enhancer := &ConditionEnhancer{}

	tests := []struct {
		name     string
		kind     string
		crdName  string
		expected bool
	}{
		{"exact match", "App", "App", true},
		{"case insensitive", "App", "app", true},
		{"no match", "Database", "Frontend", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enhancer.namesMatch(tt.kind, tt.crdName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConditionEnhancer_FindConditionsForKind(t *testing.T) {
	enhancer := &ConditionEnhancer{
		conditions: map[string][]ConditionInfo{
			"App":      {{Type: "Ready", CRDName: "App"}},
			"Database": {{Type: "Connected", CRDName: "Database"}},
		},
	}

	// Direct match
	conditions, found := enhancer.findConditionsForKind("App")
	assert.True(t, found)
	assert.Len(t, conditions, 1)
	assert.Equal(t, "Ready", conditions[0].Type)

	// No match
	_, found = enhancer.findConditionsForKind("Frontend")
	assert.False(t, found)
}

func TestConditionEnhancer_EnhanceSchemas_NoConditions(t *testing.T) {
	enhancer := &ConditionEnhancer{
		conditions: make(map[string][]ConditionInfo),
	}

	schemas := []*SchemaType{
		{Kind: "App", Group: "example.com"},
	}

	enhanced := enhancer.EnhanceSchemas(schemas)
	assert.Equal(t, schemas, enhanced) // Should return same schemas unchanged
	assert.Empty(t, enhanced[0].Conditions)
}
