//nolint:godot // ignore this entire file.
package v1alpha1

// AppConditionType defines the condition types for App CRD
type AppConditionType string

const (
	// +cty:condition:for:App
	// This condition indicates that the frontend service is ready
	FrontendReadyCond AppConditionType = "FrontendReady"

	// +cty:condition:for:App
	// This condition indicates that the backend service is ready
	BackendReadyCond AppConditionType = "BackendReady"
)

// FrontendReadyReason defines reasons for FrontendReady condition
type FrontendReadyReason string

const (
	// +cty:reason:for:App/FrontendReady
	// Frontend service is healthy and responding to requests
	FrontendHealthy FrontendReadyReason = "Healthy"

	// +cty:reason:for:App/FrontendReady
	// Frontend service is not responding or unhealthy
	FrontendUnhealthy FrontendReadyReason = "Unhealthy"

	// +cty:reason:for:App/FrontendReady
	// Frontend service is starting up
	FrontendStarting FrontendReadyReason = "Starting"
)

// BackendReadyReason defines reasons for BackendReady condition
type BackendReadyReason string

const (
	// +cty:reason:for:App/BackendReady
	// Backend database connection is established
	DatabaseConnected BackendReadyReason = "DatabaseConnected"

	// +cty:reason:for:App/BackendReady
	// Backend database connection failed
	DatabaseDisconnected BackendReadyReason = "DatabaseDisconnected"
)
