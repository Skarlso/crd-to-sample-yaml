# CRD to Sample YAML - IntelliJ Plugin

IntelliJ plugin for generating sample YAML files from Kubernetes Custom Resource Definitions.

## Prerequisites

Install the CTY binary:
```bash
go install github.com/Skarlso/crd-to-sample-yaml@latest
```

## Installation

Build the plugin and install the ZIP from `build/distributions/`:
```bash
cd intellij-plugin
./gradlew buildPlugin
```

Install via File → Settings → Plugins → Install Plugin from Disk

## Usage

Right-click on CRD YAML files to access the `CRD to Sample YAML` menu. Generate complete, minimal, or commented samples. Validate existing samples against CRD schemas.

Configure the CTY binary path and output location in File → Settings → Tools → CRD to Sample YAML.

## Development

Requires IntelliJ IDEA 2023.2+, Java 17+, and Gradle.

```bash
git clone https://github.com/Skarlso/crd-to-sample-yaml.git
cd crd-to-sample-yaml/intellij-plugin
./gradlew buildPlugin  # Build
./gradlew test         # Test
./gradlew runIde       # Run in development
```

## Troubleshooting

Install CTY if missing: `go install github.com/Skarlso/crd-to-sample-yaml@latest` or configure the full path in settings.

Context menu only appears on valid CRD files with `kind: CustomResourceDefinition` and `apiVersion: apiextensions.k8s.io/v1`.

Check IntelliJ's Event Log for error details.