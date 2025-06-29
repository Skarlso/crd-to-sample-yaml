# CRD to Sample YAML

VSCode extension for generating sample YAML files from Kubernetes Custom Resource Definitions.

## What it does

Opens CRD files and gives you buttons to generate sample YAML instances. Three flavors: complete samples, minimal (required fields only), and with comments explaining each field.

## Quick Start

Install the CTY binary first:
```bash
go install github.com/Skarlso/crd-to-sample-yaml@latest
```

Then install this extension from the marketplace. Open any CRD file and you'll see generation buttons above the `spec:` line.

## Commands

Right-click on CRD files or use the command palette:

- Generate Sample YAML (complete)
- Generate Minimal Sample YAML (required only)  
- Generate Sample YAML with Comments

## Settings

Go to VSCode settings and search "crd":

`ctyPath` - Path to the cty binary if not in PATH  
`outputLocation` - Where to save files (workspace/temp/custom)  
`autoDetectCRDs` - Show code lenses on CRD files  
`showNotifications` - Popup when files are generated  

## Example

Given this CRD:
```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  group: example.com
  names:
    kind: MyResource
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        properties:
          spec:
            properties:
              replicas:
                type: integer
                minimum: 1
              image:
                type: string
```

Generates:
```yaml
apiVersion: example.com/v1
kind: MyResource
metadata: {}
spec:
  replicas: 1
  image: string
```

## Development

Want to hack on this? Clone the repo and:

```bash
cd vscode-extension
npm install
npm run compile
```

Press F5 in VSCode to launch a development instance. The extension activates when you open YAML files.

### File Structure

```
src/
├── extension.ts      # Main entry point
├── ctyService.ts     # CTY CLI integration  
├── crdDetector.ts    # YAML parsing and CRD detection
└── crdProvider.ts    # Code lenses and commands
```

### Testing Changes

1. Make your edits
2. Run `npm run compile`
3. Press F5 to test in development host
4. Try opening CRD files and using the commands

### Publishing

```bash
npm install -g @vscode/vsce
vsce package
```

Creates a `.vsix` file you can install or publish to the marketplace.

## Troubleshooting

### CTY not found
Install it or set the `ctyPath` setting to point to your binary.

### No code Lenses
Check that `autoDetectCRDs` is enabled and you're viewing a valid CRD file.

### Generation fails
Run `cty version` to verify installation, then try `cty generate crd -c yourfile.yaml` manually.

Files are created in the same directory as the CRD by default, or in your configured output location.


# Testing

To run tests:

```bash
npm run test
```

To compile and run:

```bash
# Compile TypeScript
npm run compile

# Run linter
npm run lint

# Run tests only (after compilation)
node ./out/test/runTest.js
```
