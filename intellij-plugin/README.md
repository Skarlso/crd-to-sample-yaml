# CRD to Sample YAML - IntelliJ Plugin

IntelliJ plugin for generating sample YAML files from Kubernetes Custom Resource Definitions.

## Prerequisites

Install the `cty` binary and make sure it is on your `PATH`:
```bash
brew tap skarlso/tap
brew install crd-to-sample-yaml
```

Alternatively grab a binary from the [releases](https://github.com/Skarlso/crd-to-sample-yaml/releases),
or point the plugin at a specific binary in its settings.

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

Requires JDK 25 to build (the 2026.2 platform ships Java 25 class files); the plugin itself
targets Java 21 so it still runs on IDEs down to 2025.1.

The supported IDE range and the platform version to build against live in `gradle.properties`
(`platformVersion`, `pluginSinceBuild`, `pluginUntilBuild`) - that is the only place to bump them.

```bash
git clone https://github.com/Skarlso/crd-to-sample-yaml.git
cd crd-to-sample-yaml/intellij-plugin
./gradlew buildPlugin  # Build
./gradlew test         # Test
./gradlew runIde       # Run in development
```

## Troubleshooting

Install `cty` if missing (see Prerequisites) or configure the full path in
File | Settings | Tools | CRD to Sample YAML.

Context menu only appears on valid CRD files with `kind: CustomResourceDefinition` and an
`apiVersion` under `apiextensions.k8s.io/`.

Check IntelliJ's Event Log for error details.