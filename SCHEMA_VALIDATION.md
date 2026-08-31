# Schema Validation Feature

CTY supports two kinds of validation:

`cty validate schema` compares two *versions of a CRD* against each other to detect breaking changes.

and

`cty validate sample` checks that a *sample YAML* actually satisfies the schema of a CRD.

This is separate from tests because it adds an immediate output and can be more refined than what the test
framework is doing by showing what changed _exactly_ with some nicely formatted output. 

## Validating a sample against a CRD

```bash
# Check a sample against the schema of the CRD it belongs to
cty validate sample -c path/to/crd.yaml -s path/to/sample.yaml

# The CRD can also come from a URL
cty validate sample -u https://example.com/crd.yaml -s path/to/sample.yaml

# Ignore specific validation errors by substring
cty validate sample -c path/to/crd.yaml -s path/to/sample.yaml --ignore-errors "spec.image in body is required"
```

The version to validate against is taken from the sample's own `apiVersion`, so a CRD serving multiple
versions is handled without extra flags. Exits non-zero when the sample does not satisfy the schema.

Only `-c/--crd` and `-u/--url` are supported as CRD sources here, because validation needs the original
document rather than the parsed schema.

## Comparing versions of a schema

```bash
# Validate schema compatibility between two versions
cty validate schema -c path/to/crd.yaml --from v1alpha1 --to v1beta1

# Output formats: text (default), json, yaml
cty validate schema -c path/to/crd.yaml --from v1alpha1 --to v1beta1 -o json

# Fail build on breaking changes (useful for CI/CD)
cty validate schema -c path/to/crd.yaml --from v1alpha1 --to v1beta1 --fail-on-breaking
```

## Supported Input Sources

`validate schema` supports all the same input sources as the generate command:

- file `-c path/to/crd.yaml`
- URL `-u https://example.com/crd.yaml`
- git repository `-g https://github.com/user/repo`
- Kubernetes `-k crd-name`
- folder `-r path/to/folder`
- config file `--config path/to/config.yaml`

## Change Types

The validator detects four types of changes:

### Breaking Changes

- fields that become required break existing resources
- changing field types (e.g., string → integer)
- removing existing properties (if they have been defined by a user, it will break because it's missing)
- new or tighter constraints (minimum, maximum, pattern)

### Non-Breaking Changes
- looser constraints (lower minimum, higher maximum, removed pattern)
- making required fields optional
- removing validation constraints

### Additions
- adding new fields that aren't required
- validation rule changes that don't restrict existing valid values

## Output Formats

### Text Format (Default)
```
Schema Validation Report
=======================

CRD: TestResource
From Version: v1alpha1
To Version: v1beta1

Summary:
  Total Changes: 5
  Breaking Changes: 2
  Additions: 2
  Removals: 0

Changes:
  ⚠️ [breaking] spec.required: Field 'version' is now required
  + [addition] spec.properties.version: Property 'version' added
  ⚠️ [breaking] spec.properties.count.minimum: Minimum increased
    Old: 1
    New: 5
```

### JSON Format
```json
{
  "crdKind": "TestResource",
  "fromVersion": "v1alpha1", 
  "toVersion": "v1beta1",
  "changes": [
    {
      "type": "breaking",
      "path": "spec.required",
      "description": "Field 'version' is now required",
      "newValue": "version"
    }
  ],
  "summary": {
    "totalChanges": 5,
    "breakingChanges": 2,
    "additions": 2,
    "removals": 0
  }
}
```

### YAML Format
```yaml
crdKind: TestResource
fromVersion: v1alpha1
toVersion: v1beta1
changes:
  - type: breaking
    path: spec.required
    description: "Field 'version' is now required"
    newValue: version
summary:
  totalChanges: 5
  breakingChanges: 2
  additions: 2
  removals: 0
```

## CI/CD Integration

Use the `--fail-on-breaking` flag to make your CI/CD pipeline fail when breaking changes are detected:

```bash
# In your CI/CD pipeline
cty validate schema -c crd.yaml --from v1alpha1 --to v1beta1 --fail-on-breaking
```

This returns exit code 1 if breaking changes are found. It could be used in a couple of ways. For example, blocking
releases that break. Manual approval is required for breaking changes between versions. Automated change logs that include
this output (that would look nice).

## Examples

### Detect Breaking Changes
```bash
# Compare two versions in a multi-version CRD
cty validate schema -c manifests/crd.yaml --from v1alpha1 --to v1beta1

# Compare CRDs from different Git branches
cty validate schema -g https://github.com/user/repo --tag v1.0.0 --from v1alpha1 --to v1beta1
```

### Generate Change Reports
```bash
# Generate JSON report for automated processing
cty validate schema -c crd.yaml --from v1alpha1 --to v1beta1 -o json > changes.json

# Generate YAML report for documentation
cty validate schema -c crd.yaml --from v1alpha1 --to v1beta1 -o yaml > CHANGES.yaml
```