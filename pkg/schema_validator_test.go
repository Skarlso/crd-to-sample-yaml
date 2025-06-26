package pkg

import (
	"strings"
	"testing"

	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

func TestSchemaValidator_ValidateVersions(t *testing.T) {
	tests := []struct {
		name            string
		crd             *SchemaType
		fromVersion     string
		toVersion       string
		expectedChanges int
		expectBreaking  bool
	}{
		{
			name: "no changes between identical schemas",
			crd: &SchemaType{
				Kind: "TestResource",
				Versions: []*CRDVersion{
					{
						Name: "v1alpha1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type:     "object",
									Required: []string{"name"},
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name": {Type: "string"},
									},
								},
							},
						},
					},
					{
						Name: "v1beta1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type:     "object",
									Required: []string{"name"},
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name": {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			fromVersion:     "v1alpha1",
			toVersion:       "v1beta1",
			expectedChanges: 0,
			expectBreaking:  false,
		},
		{
			name: "breaking change - new required field",
			crd: &SchemaType{
				Kind: "TestResource",
				Versions: []*CRDVersion{
					{
						Name: "v1alpha1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type:     "object",
									Required: []string{"name"},
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name": {Type: "string"},
									},
								},
							},
						},
					},
					{
						Name: "v1beta1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type:     "object",
									Required: []string{"name", "version"},
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name":    {Type: "string"},
										"version": {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			fromVersion:     "v1alpha1",
			toVersion:       "v1beta1",
			expectedChanges: 2, // new required field + new property
			expectBreaking:  true,
		},
		{
			name: "non-breaking change - field no longer required",
			crd: &SchemaType{
				Kind: "TestResource",
				Versions: []*CRDVersion{
					{
						Name: "v1alpha1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type:     "object",
									Required: []string{"name", "version"},
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name":    {Type: "string"},
										"version": {Type: "string"},
									},
								},
							},
						},
					},
					{
						Name: "v1beta1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type:     "object",
									Required: []string{"name"},
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name":    {Type: "string"},
										"version": {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			fromVersion:     "v1alpha1",
			toVersion:       "v1beta1",
			expectedChanges: 1, // version no longer required
			expectBreaking:  false,
		},
		{
			name: "breaking change - property removed",
			crd: &SchemaType{
				Kind: "TestResource",
				Versions: []*CRDVersion{
					{
						Name: "v1alpha1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name":    {Type: "string"},
										"version": {Type: "string"},
									},
								},
							},
						},
					},
					{
						Name: "v1beta1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name": {Type: "string"},
									},
								},
							},
						},
					},
				},
			},
			fromVersion:     "v1alpha1",
			toVersion:       "v1beta1",
			expectedChanges: 1, // version property removed
			expectBreaking:  true,
		},
		{
			name: "breaking change - type changed",
			crd: &SchemaType{
				Kind: "TestResource",
				Versions: []*CRDVersion{
					{
						Name: "v1alpha1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]v1beta1.JSONSchemaProps{
										"count": {Type: "string"},
									},
								},
							},
						},
					},
					{
						Name: "v1beta1",
						Schema: &v1beta1.JSONSchemaProps{
							Type: "object",
							Properties: map[string]v1beta1.JSONSchemaProps{
								"spec": {
									Type: "object",
									Properties: map[string]v1beta1.JSONSchemaProps{
										"count": {Type: "integer"},
									},
								},
							},
						},
					},
				},
			},
			fromVersion:     "v1alpha1",
			toVersion:       "v1beta1",
			expectedChanges: 1, // type changed from string to integer
			expectBreaking:  true,
		},
	}

	validator := NewSchemaValidator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := validator.ValidateVersions(tt.crd, tt.fromVersion, tt.toVersion)
			if err != nil {
				t.Fatalf("ValidateVersions() error = %v", err)
			}

			if len(report.Changes) != tt.expectedChanges {
				t.Errorf("Expected %d changes, got %d", tt.expectedChanges, len(report.Changes))
			}

			if report.HasBreakingChanges() != tt.expectBreaking {
				t.Errorf("Expected breaking changes: %v, got: %v", tt.expectBreaking, report.HasBreakingChanges())
			}
		})
	}
}

func TestSchemaValidator_compareValidation(t *testing.T) {
	validator := NewSchemaValidator()
	
	tests := []struct {
		name            string
		from            *v1beta1.JSONSchemaProps
		to              *v1beta1.JSONSchemaProps
		expectedChanges int
		expectBreaking  bool
	}{
		{
			name: "new minimum constraint added",
			from: &v1beta1.JSONSchemaProps{Type: "integer"},
			to: &v1beta1.JSONSchemaProps{
				Type:    "integer",
				Minimum: func() *float64 { v := 5.0; return &v }(),
			},
			expectedChanges: 1,
			expectBreaking:  true,
		},
		{
			name: "minimum constraint removed",
			from: &v1beta1.JSONSchemaProps{
				Type:    "integer",
				Minimum: func() *float64 { v := 5.0; return &v }(),
			},
			to:              &v1beta1.JSONSchemaProps{Type: "integer"},
			expectedChanges: 1,
			expectBreaking:  false,
		},
		{
			name: "minimum increased (breaking)",
			from: &v1beta1.JSONSchemaProps{
				Type:    "integer",
				Minimum: func() *float64 { v := 5.0; return &v }(),
			},
			to: &v1beta1.JSONSchemaProps{
				Type:    "integer",
				Minimum: func() *float64 { v := 10.0; return &v }(),
			},
			expectedChanges: 1,
			expectBreaking:  true,
		},
		{
			name: "maximum decreased (breaking)",
			from: &v1beta1.JSONSchemaProps{
				Type:    "integer",
				Maximum: func() *float64 { v := 100.0; return &v }(),
			},
			to: &v1beta1.JSONSchemaProps{
				Type:    "integer",
				Maximum: func() *float64 { v := 50.0; return &v }(),
			},
			expectedChanges: 1,
			expectBreaking:  true,
		},
		{
			name: "pattern constraint added",
			from: &v1beta1.JSONSchemaProps{Type: "string"},
			to: &v1beta1.JSONSchemaProps{
				Type:    "string",
				Pattern: "^[a-z]+$",
			},
			expectedChanges: 1,
			expectBreaking:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := validator.compareValidation("test", tt.from, tt.to)
			
			if len(changes) != tt.expectedChanges {
				t.Errorf("Expected %d changes, got %d", tt.expectedChanges, len(changes))
			}

			hasBreaking := false
			for _, change := range changes {
				if change.Type == BreakingChange {
					hasBreaking = true
					break
				}
			}

			if hasBreaking != tt.expectBreaking {
				t.Errorf("Expected breaking changes: %v, got: %v", tt.expectBreaking, hasBreaking)
			}
		})
	}
}

func TestValidationReport_OutputText(t *testing.T) {
	report := &ValidationReport{
		CRDKind:     "TestResource",
		FromVersion: "v1alpha1",
		ToVersion:   "v1beta1",
		Changes: []Change{
			{
				Type:        BreakingChange,
				Path:        "spec.required",
				Description: "Field 'version' is now required",
				NewValue:    "version",
			},
			{
				Type:        Addition,
				Path:        "spec.properties.newField",
				Description: "Property 'newField' added",
			},
		},
		Summary: Summary{
			TotalChanges:    2,
			BreakingChanges: 1,
			Additions:       1,
			Removals:        0,
		},
	}

	var output strings.Builder
	err := report.OutputText(&output)
	if err != nil {
		t.Fatalf("OutputText() error = %v", err)
	}

	result := output.String()
	
	// Check that the output contains expected content
	expectedContent := []string{
		"Schema Validation Report",
		"CRD: TestResource",
		"From Version: v1alpha1", 
		"To Version: v1beta1",
		"Total Changes: 2",
		"Breaking Changes: 1",
		"⚠️ [breaking]",
		"+ [addition]",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected output to contain %q, but it didn't. Output:\n%s", expected, result)
		}
	}
}

func TestValidationReport_HasBreakingChanges(t *testing.T) {
	tests := []struct {
		name     string
		report   *ValidationReport
		expected bool
	}{
		{
			name: "has breaking changes",
			report: &ValidationReport{
				Summary: Summary{BreakingChanges: 1},
			},
			expected: true,
		},
		{
			name: "no breaking changes",
			report: &ValidationReport{
				Summary: Summary{BreakingChanges: 0},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.report.HasBreakingChanges(); got != tt.expected {
				t.Errorf("HasBreakingChanges() = %v, want %v", got, tt.expected)
			}
		})
	}
}