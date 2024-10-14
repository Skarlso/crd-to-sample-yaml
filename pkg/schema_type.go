package pkg

// SchemaType is a wrapper around any kind of object that provide the following:
// - kind
// - group
// - name
// - openAPIV3Schema.
type SchemaType struct {
	Schema     *JSONSchemaProps
	Versions   []*CRDVersion
	Validation *Validation
	Group      string
	Kind       string
}

// CRDVersion corresponds to a CRD version.
type CRDVersion struct {
	Name   string
	Schema *JSONSchemaProps
}

// Validation is a set of validation rules that should be applied to all versions.
type Validation struct {
	Name   string
	Schema *JSONSchemaProps
}
