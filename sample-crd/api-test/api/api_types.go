//nolint:godot // ignore this entire file.
package v1alpha1

// AppConditionType defines the condition types for App CRD
type AppConditionType string

const (
	// This condition indicates that the frontend service is ready.
	//
	// The frontend service readiness includes:
	// - Health checks passing
	// - All dependencies available
	// - Configuration loaded successfully
	//
	// When this condition is True, the frontend is operational.
	// +cty:condition:for:App
	FrontendReadyCond AppConditionType = "FrontendReady"

	// This condition indicates that the backend service is ready
	// +cty:condition:for:App
	BackendReadyCond AppConditionType = "BackendReady"
)

// FrontendReadyReason defines reasons for FrontendReady condition
type FrontendReadyReason string

const (
	// Frontend service is healthy and responding to requests.
	//
	// This means:
	// - HTTP endpoints are responding
	// - Database connectivity is established
	// - All required services are available
	// +cty:reason:for:App/FrontendReady
	FrontendHealthy FrontendReadyReason = "Healthy"

	// Frontend service is not responding or unhealthy
	// +cty:reason:for:App/FrontendReady
	FrontendUnhealthy FrontendReadyReason = "Unhealthy"

	// Frontend service is starting up
	// +cty:reason:for:App/FrontendReady
	FrontendStarting FrontendReadyReason = "Starting"
)

// BackendReadyReason defines reasons for BackendReady condition
type BackendReadyReason string

const (
	// Backend database connection is established
	// +cty:reason:for:App/BackendReady
	DatabaseConnected BackendReadyReason = "DatabaseConnected"

	// Backend database connection failed
	// +cty:reason:for:App/BackendReady
	DatabaseDisconnected BackendReadyReason = "DatabaseDisconnected"
)
