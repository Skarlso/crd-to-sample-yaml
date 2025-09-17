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

func TestConditionParser_RecursiveDirectoryParsing(t *testing.T) {
	// Create a temporary directory structure with nested Go files
	tempDir := t.TempDir()

	// Create subdirectories
	v1Dir := filepath.Join(tempDir, "v1")
	v2Dir := filepath.Join(tempDir, "v2")
	nestedDir := filepath.Join(tempDir, "v1", "nested")

	require.NoError(t, os.MkdirAll(v1Dir, 0755))
	require.NoError(t, os.MkdirAll(v2Dir, 0755))
	require.NoError(t, os.MkdirAll(nestedDir, 0755))

	// Create test Go files in different directories
	v1File := `package v1

const (
	// +cty:condition:for:AppV1
	// V1 specific condition
	V1ReadyCond = "V1Ready"
)

const (
	// +cty:reason:for:AppV1/V1ReadyCond
	// V1 is operational
	V1Operational = "Operational"
)
`

	v2File := `package v2

const (
	// +cty:condition:for:AppV2
	// V2 specific condition
	V2ReadyCond = "V2Ready"
)

const (
	// +cty:reason:for:AppV2/V2ReadyCond
	// V2 is running
	V2Running = "Running"
)
`

	nestedFile := `package nested

const (
	// +cty:condition:for:NestedApp
	// Nested condition
	NestedCond = "NestedReady"
)

const (
	// +cty:reason:for:NestedApp/NestedCond
	// Nested is ready
	NestedReady = "Ready"
)
`

	// Write test files to different directories
	require.NoError(t, os.WriteFile(filepath.Join(v1Dir, "v1_types.go"), []byte(v1File), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(v2Dir, "v2_types.go"), []byte(v2File), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(nestedDir, "nested_types.go"), []byte(nestedFile), 0644))

	// Add a non-Go file to ensure it's ignored
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "README.md"), []byte("# Documentation"), 0644))

	// Parse the directory recursively
	parser := NewConditionParser()
	err := parser.ParseGoFiles(tempDir)
	require.NoError(t, err)

	// Get conditions
	conditions := parser.GetConditions()

	// Verify all conditions from different directories were parsed
	assert.Len(t, conditions, 3, "Should have conditions from all subdirectories")

	// Verify V1 conditions
	v1Conditions, exists := conditions["AppV1"]
	require.True(t, exists, "AppV1 conditions should exist")
	require.Len(t, v1Conditions, 1, "Should have 1 AppV1 condition")
	assert.Equal(t, "V1ReadyCond", v1Conditions[0].Type)
	assert.Equal(t, "V1 specific condition", v1Conditions[0].Description)
	assert.Len(t, v1Conditions[0].Reasons, 1, "Should have 1 reason")
	assert.Equal(t, "V1Operational", v1Conditions[0].Reasons[0].Name)

	// Verify V2 conditions
	v2Conditions, exists := conditions["AppV2"]
	require.True(t, exists, "AppV2 conditions should exist")
	require.Len(t, v2Conditions, 1, "Should have 1 AppV2 condition")
	assert.Equal(t, "V2ReadyCond", v2Conditions[0].Type)
	assert.Equal(t, "V2 specific condition", v2Conditions[0].Description)
	assert.Len(t, v2Conditions[0].Reasons, 1, "Should have 1 reason")
	assert.Equal(t, "V2Running", v2Conditions[0].Reasons[0].Name)

	// Verify Nested conditions
	nestedConditions, exists := conditions["NestedApp"]
	require.True(t, exists, "NestedApp conditions should exist")
	require.Len(t, nestedConditions, 1, "Should have 1 NestedApp condition")
	assert.Equal(t, "NestedCond", nestedConditions[0].Type)
	assert.Equal(t, "Nested condition", nestedConditions[0].Description)
	assert.Len(t, nestedConditions[0].Reasons, 1, "Should have 1 reason")
	assert.Equal(t, "NestedReady", nestedConditions[0].Reasons[0].Name)
}