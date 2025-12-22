package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"gopkg.in/yaml.v3"

	"github.com/Skarlso/crd-to-sample-yaml/v1beta1"
)

// ChangeType represents the type of schema change detected.
type ChangeType string

const (
	BreakingChange    ChangeType = "breaking"
	NonBreakingChange ChangeType = "non-breaking"
	Addition          ChangeType = "addition"
	Removal           ChangeType = "removal"
)

// Change represents a single schema change.
type Change struct {
	Type        ChangeType `json:"type"               yaml:"type"`
	Path        string     `json:"path"               yaml:"path"`
	Description string     `json:"description"        yaml:"description"`
	OldValue    string     `json:"oldValue,omitempty" yaml:"oldValue,omitempty"`
	NewValue    string     `json:"newValue,omitempty" yaml:"newValue,omitempty"`
}

// ValidationReport contains the results of schema validation.
type ValidationReport struct {
	CRDKind     string   `json:"crdKind"     yaml:"crdKind"`
	FromVersion string   `json:"fromVersion" yaml:"fromVersion"`
	ToVersion   string   `json:"toVersion"   yaml:"toVersion"`
	Changes     []Change `json:"changes"     yaml:"changes"`
	Summary     Summary  `json:"summary"     yaml:"summary"`
}

// Summary provides an overview of changes.
type Summary struct {
	TotalChanges    int `json:"totalChanges"    yaml:"totalChanges"`
	BreakingChanges int `json:"breakingChanges" yaml:"breakingChanges"`
	Additions       int `json:"additions"       yaml:"additions"`
	Removals        int `json:"removals"        yaml:"removals"`
}

// HasBreakingChanges returns true if the report contains any breaking changes.
func (r *ValidationReport) HasBreakingChanges() bool {
	return r.Summary.BreakingChanges > 0
}

// OutputJSON writes the validation report as JSON.
func (r *ValidationReport) OutputJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	return encoder.Encode(r)
}

// OutputYAML writes the validation report as YAML.
func (r *ValidationReport) OutputYAML(w io.Writer) error {
	encoder := yaml.NewEncoder(w)

	defer func() {
		_ = encoder.Close()
	}()

	return encoder.Encode(r)
}

// OutputText writes the validation report as human-readable text.
func (r *ValidationReport) OutputText(w io.Writer) error {
	wr := &writer{}
	wr.write(w, "Schema Validation Report\n")
	wr.write(w, "=======================\n\n")
	wr.write(w, fmt.Sprintf("CRD: %s\n", r.CRDKind))
	wr.write(w, fmt.Sprintf("From Version: %s\n", r.FromVersion))
	wr.write(w, fmt.Sprintf("To Version: %s\n\n", r.ToVersion))

	wr.write(w, "Summary:\n")
	wr.write(w, fmt.Sprintf("  Total Changes: %d\n", r.Summary.TotalChanges))
	wr.write(w, fmt.Sprintf("  Breaking Changes: %d\n", r.Summary.BreakingChanges))
	wr.write(w, fmt.Sprintf("  Additions: %d\n", r.Summary.Additions))
	wr.write(w, fmt.Sprintf("  Removals: %d\n\n", r.Summary.Removals))

	if len(r.Changes) == 0 {
		wr.write(w, "No changes detected.\n")

		return nil
	}

	wr.write(w, "Changes:\n")

	for _, change := range r.Changes {
		symbol := getChangeSymbol(change.Type)
		wr.write(w, fmt.Sprintf("  %s [%s] %s: %s\n", symbol, change.Type, change.Path, change.Description))

		if change.OldValue != "" {
			wr.write(w, fmt.Sprintf("    Old: %s\n", change.OldValue))
		}

		if change.NewValue != "" {
			wr.write(w, fmt.Sprintf("    New: %s\n", change.NewValue))
		}
	}

	if wr.err != nil {
		return fmt.Errorf("failed to write report: %w", wr.err)
	}

	return nil
}

func getChangeSymbol(changeType ChangeType) string {
	switch changeType {
	case BreakingChange:
		return "⚠️"
	case Addition:
		return "+"
	case Removal:
		return "-"
	default:
		return "~"
	}
}

// SchemaValidator validates schema compatibility between versions.
type SchemaValidator struct{}

// NewSchemaValidator creates a new schema validator.
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{}
}

// ValidateVersions compares two versions of a CRD schema and reports changes.
func (v *SchemaValidator) ValidateVersions(crd *SchemaType, fromVersion, toVersion string) (*ValidationReport, error) {
	fromSchema, err := v.findVersionSchema(crd, fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to find from version %s: %w", fromVersion, err)
	}

	toSchema, err := v.findVersionSchema(crd, toVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to find to version %s: %w", toVersion, err)
	}

	changes := v.compareSchemas("spec", fromSchema, toSchema)

	report := &ValidationReport{
		CRDKind:     crd.Kind,
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Changes:     changes,
		Summary:     v.calculateSummary(changes),
	}

	return report, nil
}

func (v *SchemaValidator) findVersionSchema(crd *SchemaType, version string) (*v1beta1.JSONSchemaProps, error) {
	if version == "" && len(crd.Versions) > 0 {
		return crd.Versions[0].Schema, nil
	}

	for _, ver := range crd.Versions {
		if ver.Name == version {
			return ver.Schema, nil
		}
	}

	if crd.Validation != nil && (version == "" || version == crd.Validation.Name) {
		return crd.Validation.Schema, nil
	}

	return nil, fmt.Errorf("version %s not found", version)
}

func (v *SchemaValidator) compareSchemas(basePath string, from, to *v1beta1.JSONSchemaProps) []Change {
	var changes []Change

	if from == nil && to == nil {
		return changes
	}

	if from == nil {
		changes = append(changes, Change{
			Type:        Addition,
			Path:        basePath,
			Description: "New schema added",
		})

		return changes
	}

	if to == nil {
		changes = append(changes, Change{
			Type:        Removal,
			Path:        basePath,
			Description: "Schema removed",
		})

		return changes
	}

	// Compare types
	if from.Type != to.Type {
		changes = append(changes, Change{
			Type:        BreakingChange,
			Path:        basePath + ".type",
			Description: "Type changed",
			OldValue:    from.Type,
			NewValue:    to.Type,
		})
	}

	// Compare required fields
	changes = append(changes, v.compareRequired(basePath, from.Required, to.Required)...)

	// Compare properties
	changes = append(changes, v.compareProperties(basePath, from.Properties, to.Properties)...)

	// Compare validation constraints
	changes = append(changes, v.compareValidation(basePath, from, to)...)

	return changes
}

func (v *SchemaValidator) compareRequired(basePath string, fromRequired, toRequired []string) []Change {
	var changes []Change

	// Check for newly required fields (breaking change)
	for _, req := range toRequired {
		if !slices.Contains(fromRequired, req) {
			changes = append(changes, Change{
				Type:        BreakingChange,
				Path:        basePath + ".required",
				Description: fmt.Sprintf("Field '%s' is now required", req),
				NewValue:    req,
			})
		}
	}

	// Check for fields no longer required (non-breaking change)
	for _, req := range fromRequired {
		if !slices.Contains(toRequired, req) {
			changes = append(changes, Change{
				Type:        NonBreakingChange,
				Path:        basePath + ".required",
				Description: fmt.Sprintf("Field '%s' is no longer required", req),
				OldValue:    req,
			})
		}
	}

	return changes
}

func (v *SchemaValidator) compareProperties(basePath string, fromProps, toProps map[string]v1beta1.JSONSchemaProps) []Change {
	var changes []Change

	// Check for removed properties (breaking change)
	for propName := range fromProps {
		if _, exists := toProps[propName]; !exists {
			changes = append(changes, Change{
				Type:        BreakingChange,
				Path:        fmt.Sprintf("%s.properties.%s", basePath, propName),
				Description: fmt.Sprintf("Property '%s' removed", propName),
			})
		}
	}

	// Check for added properties (addition)
	for propName := range toProps {
		if _, exists := fromProps[propName]; !exists {
			changes = append(changes, Change{
				Type:        Addition,
				Path:        fmt.Sprintf("%s.properties.%s", basePath, propName),
				Description: fmt.Sprintf("Property '%s' added", propName),
			})
		}
	}

	// Compare existing properties recursively
	for propName, fromProp := range fromProps {
		if toProp, exists := toProps[propName]; exists {
			propPath := fmt.Sprintf("%s.properties.%s", basePath, propName)
			changes = append(changes, v.compareSchemas(propPath, &fromProp, &toProp)...)
		}
	}

	return changes
}

func (v *SchemaValidator) compareValidation(basePath string, from, to *v1beta1.JSONSchemaProps) []Change {
	var changes []Change

	// Compare minimum values
	switch {
	case from.Minimum != nil && to.Minimum != nil:
		if *from.Minimum != *to.Minimum {
			if *to.Minimum > *from.Minimum {
				changes = append(changes, Change{
					Type:        BreakingChange,
					Path:        basePath + ".minimum",
					Description: "Minimum value increased",
					OldValue:    fmt.Sprintf("%.0f", *from.Minimum),
					NewValue:    fmt.Sprintf("%.0f", *to.Minimum),
				})
			} else {
				changes = append(changes, Change{
					Type:        NonBreakingChange,
					Path:        basePath + ".minimum",
					Description: "Minimum value decreased",
					OldValue:    fmt.Sprintf("%.0f", *from.Minimum),
					NewValue:    fmt.Sprintf("%.0f", *to.Minimum),
				})
			}
		}
	case from.Minimum == nil && to.Minimum != nil:
		changes = append(changes, Change{
			Type:        BreakingChange,
			Path:        basePath + ".minimum",
			Description: "New minimum constraint added",
			NewValue:    fmt.Sprintf("%.0f", *to.Minimum),
		})
	case from.Minimum != nil && to.Minimum == nil:
		changes = append(changes, Change{
			Type:        NonBreakingChange,
			Path:        basePath + ".minimum",
			Description: "Minimum constraint removed",
			OldValue:    fmt.Sprintf("%.0f", *from.Minimum),
		})
	}

	// Compare maximum values
	switch {
	case from.Maximum != nil && to.Maximum != nil:
		if *from.Maximum != *to.Maximum {
			if *to.Maximum < *from.Maximum {
				changes = append(changes, Change{
					Type:        BreakingChange,
					Path:        basePath + ".maximum",
					Description: "Maximum value decreased",
					OldValue:    fmt.Sprintf("%.0f", *from.Maximum),
					NewValue:    fmt.Sprintf("%.0f", *to.Maximum),
				})
			} else {
				changes = append(changes, Change{
					Type:        NonBreakingChange,
					Path:        basePath + ".maximum",
					Description: "Maximum value increased",
					OldValue:    fmt.Sprintf("%.0f", *from.Maximum),
					NewValue:    fmt.Sprintf("%.0f", *to.Maximum),
				})
			}
		}
	case from.Maximum == nil && to.Maximum != nil:
		changes = append(changes, Change{
			Type:        BreakingChange,
			Path:        basePath + ".maximum",
			Description: "New maximum constraint added",
			NewValue:    fmt.Sprintf("%.0f", *to.Maximum),
		})
	case from.Maximum != nil && to.Maximum == nil:
		changes = append(changes, Change{
			Type:        NonBreakingChange,
			Path:        basePath + ".maximum",
			Description: "Maximum constraint removed",
			OldValue:    fmt.Sprintf("%.0f", *from.Maximum),
		})
	}

	// Compare pattern validation
	if from.Pattern != to.Pattern {
		switch {
		case from.Pattern == "" && to.Pattern != "":
			changes = append(changes, Change{
				Type:        BreakingChange,
				Path:        basePath + ".pattern",
				Description: "New pattern constraint added",
				NewValue:    to.Pattern,
			})
		case from.Pattern != "" && to.Pattern == "":
			changes = append(changes, Change{
				Type:        NonBreakingChange,
				Path:        basePath + ".pattern",
				Description: "Pattern constraint removed",
				OldValue:    from.Pattern,
			})
		case from.Pattern != "" && to.Pattern != "":
			changes = append(changes, Change{
				Type:        BreakingChange,
				Path:        basePath + ".pattern",
				Description: "Pattern constraint changed",
				OldValue:    from.Pattern,
				NewValue:    to.Pattern,
			})
		}
	}

	return changes
}

func (v *SchemaValidator) calculateSummary(changes []Change) Summary {
	summary := Summary{
		TotalChanges: len(changes),
	}

	for _, change := range changes {
		switch change.Type {
		case BreakingChange:
			summary.BreakingChanges++
		case Addition:
			summary.Additions++
		case Removal:
			summary.Removals++
		}
	}

	return summary
}
