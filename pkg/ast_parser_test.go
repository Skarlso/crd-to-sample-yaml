package pkg

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConditionParser_ParseGoFiles(t *testing.T) {
	// Create a temporary directory with test Go files
	tempDir := t.TempDir()

	// Create a test Go file with condition annotations
	testFile1 := `package types

import "k8s.io/apimachinery/pkg/runtime"

// AppConditionType defines the condition types
type AppConditionType string

const (
	// +cty:condition:for:App
	// This condition indicates that the frontend is ready
	FrontendReadyCond AppConditionType = "FrontendReady"

	// +cty:condition:for:App
	// This condition indicates that the backend is ready
	BackendReadyCond AppConditionType = "BackendReady"
)

// FrontendReadyReason defines reasons for FrontendReady condition
type FrontendReadyReason string

const (
	// +cty:reason:for:App/FrontendReady
	// Frontend service is healthy and responding
	FrontendReady FrontendReadyReason = "Ready"

	// +cty:reason:for:App/FrontendReady
	// Frontend service is not responding
	FrontendNotReady FrontendReadyReason = "NotReady"
)

// BackendReadyReason defines reasons for BackendReady condition
type BackendReadyReason string

const (
	// +cty:reason:for:App/BackendReady
	// Backend database connection is established
	BackendReady BackendReadyReason = "DatabaseConnected"

	// +cty:reason:for:App/BackendReady
	// Backend database connection failed
	BackendNotReady BackendReadyReason = "DatabaseDisconnected"
)
`

	testFile2 := `package types

// DatabaseConditionType defines database-related conditions
type DatabaseConditionType string

const (
	// +cty:condition:for:Database
	// Connection status to the database
	DBConnectionCond DatabaseConditionType = "Connection"
)

// DatabaseConnectionReason defines reasons for database connection
type DatabaseConnectionReason string

const (
	// +cty:reason:for:Database/Connection
	// Successfully connected to database
	Connected DatabaseConnectionReason = "Connected"

	// +cty:reason:for:Database/Connection
	// Failed to connect to database
	Disconnected DatabaseConnectionReason = "Disconnected"
)
`

	// Write test files
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "app_types.go"), []byte(testFile1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "db_types.go"), []byte(testFile2), 0644))

	// Parse the files
	parser := NewConditionParser()
	err := parser.ParseGoFiles(tempDir)
	require.NoError(t, err)

	// Get conditions
	conditions := parser.GetConditions()

	// Verify App conditions
	appConditions, exists := conditions["App"]
	require.True(t, exists, "App conditions should exist")
	require.Len(t, appConditions, 2, "Should have 2 App conditions")

	// Verify FrontendReady condition
	var frontendCond *ConditionInfo
	for i := range appConditions {
		if appConditions[i].Type == "FrontendReadyCond" {
			frontendCond = &appConditions[i]
			break
		}
	}
	require.NotNil(t, frontendCond, "FrontendReady condition should exist")
	assert.Equal(t, "App", frontendCond.CRDName)
	assert.Equal(t, "FrontendReadyCond", frontendCond.Type)
	assert.Equal(t, "This condition indicates that the frontend is ready", frontendCond.Description)
	assert.Len(t, frontendCond.Reasons, 2, "Should have 2 reasons")

	// Verify reasons for FrontendReady
	reasonNames := make([]string, len(frontendCond.Reasons))
	for i, reason := range frontendCond.Reasons {
		reasonNames[i] = reason.Name
	}
	assert.Contains(t, reasonNames, "FrontendReady")
	assert.Contains(t, reasonNames, "FrontendNotReady")

	// Find and verify specific reason
	var readyReason *ReasonInfo
	for i := range frontendCond.Reasons {
		if frontendCond.Reasons[i].Name == "FrontendReady" {
			readyReason = &frontendCond.Reasons[i]
			break
		}
	}
	require.NotNil(t, readyReason, "Ready reason should exist")
	assert.Equal(t, "Ready", readyReason.Value)
	assert.Equal(t, "Frontend service is healthy and responding", readyReason.Description)

	// Verify Database conditions
	dbConditions, exists := conditions["Database"]
	require.True(t, exists, "Database conditions should exist")
	require.Len(t, dbConditions, 1, "Should have 1 Database condition")

	dbCond := dbConditions[0]
	assert.Equal(t, "Database", dbCond.CRDName)
	assert.Equal(t, "DBConnectionCond", dbCond.Type)
	assert.Equal(t, "Connection status to the database", dbCond.Description)
	assert.Len(t, dbCond.Reasons, 2, "Should have 2 reasons")
}

func TestConditionParser_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	parser := NewConditionParser()
	err := parser.ParseGoFiles(tempDir)
	require.NoError(t, err)

	conditions := parser.GetConditions()
	assert.Empty(t, conditions, "Should have no conditions for empty directory")
}

func TestConditionParser_NoAnnotations(t *testing.T) {
	tempDir := t.TempDir()

	// Create a Go file without any annotations
	testFile := `package types

type SomeType string

const (
	Value1 SomeType = "value1"
	Value2 SomeType = "value2"
)
`

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "types.go"), []byte(testFile), 0644))

	parser := NewConditionParser()
	err := parser.ParseGoFiles(tempDir)
	require.NoError(t, err)

	conditions := parser.GetConditions()
	assert.Empty(t, conditions, "Should have no conditions when no annotations present")
}

func TestConditionParser_InvalidGoFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create an invalid Go file
	invalidFile := `this is not valid go code`
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "invalid.go"), []byte(invalidFile), 0644))

	parser := NewConditionParser()
	err := parser.ParseGoFiles(tempDir)
	require.Error(t, err)
}

func TestConditionParser_OrphanedReasons(t *testing.T) {
	tempDir := t.TempDir()

	// Create a Go file with reasons but no matching conditions
	testFile := `package types

type OrphanReason string

const (
	// +cty:reason:for:NonExistent/SomeCondition
	// This reason has no matching condition
	OrphanedReason OrphanReason = "Orphaned"
)
`

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "orphan.go"), []byte(testFile), 0644))

	parser := NewConditionParser()
	err := parser.ParseGoFiles(tempDir)
	require.NoError(t, err)

	conditions := parser.GetConditions()
	assert.Empty(t, conditions, "Should have no conditions when only orphaned reasons exist")
}

func TestConditionParser_NonExistentDirectory(t *testing.T) {
	parser := NewConditionParser()
	err := parser.ParseGoFiles("/non/existent/path")
	require.Error(t, err)
}