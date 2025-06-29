# CRD to Sample YAML Extension Development

## Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Skarlso/crd-to-sample-yaml.git
   cd crd-to-sample-yaml/vscode-extension
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Build the extension**
   ```bash
   npm run compile
   ```

4. **Debug the extension**
   - Open the extension folder in VSCode
   - Press `F5` to launch a new Extension Development Host window
   - Test your changes in the new window

## Testing

1. **Install test dependencies**
   ```bash
   npm install
   ```

2. **Run tests**
   ```bash
   npm test
   ```

## Packaging

1. **Install VSCE (Visual Studio Code Extension manager)**
   ```bash
   npm install -g @vscode/vsce
   ```

2. **Package the extension**
   ```bash
   npm run package
   ```

This creates a `.vsix` file that can be installed manually or published to the marketplace.
