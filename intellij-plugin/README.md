# CRD to Sample YAML - IntelliJ Plugin

IntelliJ IDEA plugin for generating sample YAML files from Kubernetes Custom Resource Definitions using the CTY tool.

## Features

- **CRD Detection**: Automatically detects CRD files in your project
- **Sample Generation**: Generate complete, minimal, or commented sample YAML files
- **Validation**: Validate sample YAML files against CRD schemas
- **Context Menu Integration**: Right-click on CRD files to access all features
- **Configurable**: Settings panel to configure CTY binary path and output options

## Prerequisites

Install the CTY binary:
```bash
go install github.com/Skarlso/crd-to-sample-yaml@latest
```

## Installation

1. Build the plugin:
   ```bash
   cd intellij-plugin
   ./gradlew buildPlugin
   ```

2. Install the generated `.zip` file from `build/distributions/` via:
   - **File → Settings → Plugins → Install Plugin from Disk**

## Usage

### Generate Samples
1. Right-click on any CRD YAML file
2. Select **CRD to Sample YAML** from the context menu
3. Choose your generation option:
   - **Generate Sample YAML** - Complete sample with all fields
   - **Generate Minimal Sample YAML** - Required fields only
   - **Generate Sample YAML with Comments** - Sample with field descriptions

### Validate Samples
1. Right-click on a CRD file
2. Select **CRD to Sample YAML → Validate Sample Against CRD**
3. Choose the sample YAML file to validate

### Configure Settings
1. Go to **File → Settings → Tools → CRD to Sample YAML**
2. Configure:
   - **CTY Binary Path**: Path to the cty executable (leave empty to use PATH)
   - **Output Location**: Where to save generated files
   - **Show Notifications**: Enable/disable success/error notifications

## Development

### Prerequisites
- IntelliJ IDEA 2023.2+
- Java 17+
- Gradle

### Setup
```bash
git clone https://github.com/Skarlso/crd-to-sample-yaml.git
cd crd-to-sample-yaml/intellij-plugin
./gradlew buildPlugin
```

### Testing
```bash
./gradlew test
```

### Running in Development
```bash
./gradlew runIde
```

### Project Structure
```
src/main/java/com/skarlso/crdtosampleyaml/
├── actions/               # Context menu actions
│   ├── BaseAction.java
│   ├── GenerateSampleAction.java
│   ├── GenerateMinimalSampleAction.java
│   ├── GenerateSampleWithCommentsAction.java
│   └── ValidateSampleAction.java
├── services/              # Core services
│   ├── CrdDetectorService.java
│   └── CtyService.java
└── settings/              # Plugin configuration
    ├── CtySettings.java
    └── CtyConfigurable.java
```

## Troubleshooting

### CTY Not Found
- Install CTY: `go install github.com/Skarlso/crd-to-sample-yaml@latest`
- Or configure the full path in plugin settings

### No Context Menu Options
- Ensure you're right-clicking on a valid CRD YAML file
- CRD files must have `kind: CustomResourceDefinition` and `apiVersion: apiextensions.k8s.io/v1`

### Generation Fails
- Check that CTY is properly installed: `cty version`
- Verify the CRD file is valid
- Check IntelliJ's Event Log for detailed error messages

## Building for Distribution

```bash
./gradlew buildPlugin
```

The plugin ZIP file will be created in `build/distributions/`

## License

This plugin is part of the CRD to Sample YAML project and follows the same license terms.