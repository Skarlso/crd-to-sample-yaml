import * as vscode from 'vscode';
import { CRDProvider } from './crdProvider';
import { CtyService } from './ctyService';
import { CRDDetector } from './crdDetector';

export function activate(context: vscode.ExtensionContext) {
    console.log('CRD to Sample YAML extension is now active');

    const ctyService = new CtyService();
    const crdDetector = new CRDDetector();
    const crdProvider = new CRDProvider(ctyService, crdDetector);

    // Register commands
    const generateSampleCommand = vscode.commands.registerCommand(
        'crdToSampleYaml.generateSample',
        async (uri?: vscode.Uri) => {
            await crdProvider.generateSample(uri, { minimal: false, comments: false });
        }
    );

    const generateMinimalSampleCommand = vscode.commands.registerCommand(
        'crdToSampleYaml.generateMinimalSample',
        async (uri?: vscode.Uri) => {
            await crdProvider.generateSample(uri, { minimal: true, comments: false });
        }
    );

    const generateSampleWithCommentsCommand = vscode.commands.registerCommand(
        'crdToSampleYaml.generateSampleWithComments',
        async (uri?: vscode.Uri) => {
            await crdProvider.generateSample(uri, { minimal: false, comments: true });
        }
    );

    const validateSampleCommand = vscode.commands.registerCommand(
        'crdToSampleYaml.validateSample',
        async (uri?: vscode.Uri) => {
            await crdProvider.validateSample(uri);
        }
    );

    // Register code lens provider for CRD files
    const codeLensProvider = vscode.languages.registerCodeLensProvider(
        { language: 'yaml' },
        crdProvider
    );

    // Register hover provider for enhanced tooltips
    const hoverProvider = vscode.languages.registerHoverProvider(
        { language: 'yaml' },
        crdProvider
    );

    // Watch for file changes to update code lenses
    const fileWatcher = vscode.workspace.createFileSystemWatcher('**/*.{yaml,yml}');
    fileWatcher.onDidChange(() => {
        vscode.commands.executeCommand('vscode.executeCodeLensProvider');
    });

    context.subscriptions.push(
        generateSampleCommand,
        generateMinimalSampleCommand,
        generateSampleWithCommentsCommand,
        validateSampleCommand,
        codeLensProvider,
        hoverProvider,
        fileWatcher
    );

    // Show welcome message on first activation
    const hasShownWelcome = context.globalState.get('hasShownWelcome', false);
    if (!hasShownWelcome) {
        vscode.window.showInformationMessage(
            'CRD to Sample YAML extension activated! Right-click on CRD files to generate samples.',
            'Got it!'
        ).then(() => {
            context.globalState.update('hasShownWelcome', true);
        });
    }
}

export function deactivate() {
    console.log('CRD to Sample YAML extension is now deactivated');
}