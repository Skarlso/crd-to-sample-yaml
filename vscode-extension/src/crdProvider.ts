import * as vscode from 'vscode';
import * as path from 'path';
import { CtyService, GenerationOptions } from './ctyService';
import { CRDDetector, CRDInfo } from './crdDetector';

export class CRDProvider implements vscode.CodeLensProvider, vscode.HoverProvider {
    private _onDidChangeCodeLenses: vscode.EventEmitter<void> = new vscode.EventEmitter<void>();
    public readonly onDidChangeCodeLenses: vscode.Event<void> = this._onDidChangeCodeLenses.event;

    constructor(
        private ctyService: CtyService,
        private crdDetector: CRDDetector
    ) {}

    async provideCodeLenses(document: vscode.TextDocument): Promise<vscode.CodeLens[]> {
        const config = vscode.workspace.getConfiguration('crdToSampleYaml');
        const autoDetect = config.get<boolean>('autoDetectCRDs', true);
        
        if (!autoDetect) {
            return [];
        }

        const crdInfo = await this.crdDetector.detectCRD(document);
        if (!crdInfo?.isCRD) {
            return [];
        }

        const codeLenses: vscode.CodeLens[] = [];
        const lineNumbers = this.crdDetector.getCRDLineNumbers(document);

        // code-lens
        if (lineNumbers.specLine >= 0) {
            const range = new vscode.Range(lineNumbers.specLine, 0, lineNumbers.specLine, 0);
            
            codeLenses.push(new vscode.CodeLens(range, {
                title: "🔧 Generate Sample",
                command: 'crdToSampleYaml.generateSample',
                arguments: [document.uri]
            }));
            
            codeLenses.push(new vscode.CodeLens(range, {
                title: "📝 Generate Minimal",
                command: 'crdToSampleYaml.generateMinimalSample',
                arguments: [document.uri]
            }));
            
            codeLenses.push(new vscode.CodeLens(range, {
                title: "💬 Generate with Comments",
                command: 'crdToSampleYaml.generateSampleWithComments',
                arguments: [document.uri]
            }));
        }
        
        return codeLenses;
    }

    async provideHover(document: vscode.TextDocument, position: vscode.Position): Promise<vscode.Hover | undefined> {
        const crdInfo = await this.crdDetector.detectCRD(document);
        if (!crdInfo?.isCRD) {
            return undefined;
        }

        const line = document.lineAt(position.line);
        const lineText = line.text.trim();
        
        if (lineText.startsWith('kind:') && lineText.includes('CustomResourceDefinition')) {
            const markdown = new vscode.MarkdownString();
            markdown.appendMarkdown(`**Custom Resource Definition**\n\n`);
            markdown.appendMarkdown(`• **Kind**: ${crdInfo.kind}\n`);
            markdown.appendMarkdown(`• **Group**: ${crdInfo.group}\n`);
            markdown.appendMarkdown(`• **Versions**: ${crdInfo.versions.join(', ')}\n\n`);
            markdown.appendMarkdown(`Right-click to generate sample YAML files.`);
            
            return new vscode.Hover(markdown);
        }
        
        return undefined;
    }

    async generateSample(uri?: vscode.Uri, options?: GenerationOptions): Promise<void> {
        const targetUri = uri || this.getActiveDocumentUri();
        if (!targetUri) {
            vscode.window.showErrorMessage('No CRD file selected or active');
            return;
        }

        // Check if cty is available
        const ctyAvailable = await this.ctyService.checkCtyAvailable();
        if (!ctyAvailable) {
            const result = await vscode.window.showErrorMessage(
                'CTY binary not found. Please install crd-to-sample-yaml or configure the path.',
                'Configure Path',
                'Install Instructions'
            );
            
            if (result === 'Configure Path') {
                vscode.commands.executeCommand('workbench.action.openSettings', 'crdToSampleYaml.ctyPath');
            } else if (result === 'Install Instructions') {
                vscode.env.openExternal(vscode.Uri.parse('https://github.com/Skarlso/crd-to-sample-yaml#getting-started'));
            }
            return;
        }

        // Validate it's a CRD file
        if (!this.crdDetector.isValidCRDFile(targetUri)) {
            vscode.window.showErrorMessage('Selected file is not a valid YAML file');
            return;
        }

        const document = await vscode.workspace.openTextDocument(targetUri);
        const crdInfo = await this.crdDetector.detectCRD(document);
        
        if (!crdInfo?.isCRD) {
            vscode.window.showErrorMessage('Selected file does not appear to be a CRD');
            return;
        }

        const defaultOptions: GenerationOptions = {
            minimal: false,
            comments: false,
            ...options
        };

        try {
            // Show progress
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: "Generating sample YAML...",
                cancellable: false
            }, async (progress) => {
                progress.report({ increment: 30, message: "Processing CRD..." });
                
                // Add output path for custom location
                const config = vscode.workspace.getConfiguration('crdToSampleYaml');
                const outputLocation = config.get<string>('outputLocation', 'workspace');
                const customPath = config.get<string>('customOutputPath', '');
                
                if (outputLocation === 'custom' && customPath) {
                    defaultOptions.output = customPath;
                }
                
                // Actually generate the sample
                console.log(`Generating sample for: ${targetUri.fsPath}`, defaultOptions);
                await this.ctyService.generateSample(targetUri.fsPath, defaultOptions);
                
                progress.report({ increment: 70, message: "Opening generated file..." });
                
                // Get the expected output path and open the file
                const samplePath = await this.ctyService.getGeneratedSamplePath(targetUri.fsPath, defaultOptions);
                console.log(`Expected sample path: ${samplePath}`);
                
                // Check if a file exists before trying to open it
                const fs = require('fs');
                if (!fs.existsSync(samplePath)) {
                    // List what files are actually in the directory
                    const outputDir = path.dirname(samplePath);
                    try {
                        const files = fs.readdirSync(outputDir);
                        console.log(`Files in output directory ${outputDir}:`, files);
                        throw new Error(`Generated sample file not found at: ${samplePath}\nFiles in directory: ${files.join(', ')}`);
                    } catch (dirError) {
                        throw new Error(`Generated sample file not found at: ${samplePath}\nCannot read directory: ${outputDir}`);
                    }
                }
                
                const sampleUri = vscode.Uri.file(samplePath);
                const sampleDocument = await vscode.workspace.openTextDocument(sampleUri);
                await vscode.window.showTextDocument(sampleDocument, vscode.ViewColumn.Beside);
                
                const showNotifications = config.get<boolean>('showNotifications', true);
                
                if (showNotifications) {
                    const fileName = path.basename(samplePath);
                    vscode.window.showInformationMessage(
                        `Sample YAML generated: ${fileName}`,
                        'Open Folder'
                    ).then(selection => {
                        if (selection === 'Open Folder') {
                            vscode.commands.executeCommand('revealFileInOS', sampleUri);
                        }
                    });
                }
            });
        } catch (error) {
            vscode.window.showErrorMessage(`Failed to generate sample: ${error instanceof Error ? error.message : 'Unknown error'}`);
        }
    }

    async validateSample(uri?: vscode.Uri): Promise<void> {
        const targetUri = uri || this.getActiveDocumentUri();
        if (!targetUri) {
            vscode.window.showErrorMessage('No file selected or active');
            return;
        }

        vscode.window.showInformationMessage('Sample validation feature coming soon!');
    }

    private getActiveDocumentUri(): vscode.Uri | undefined {
        return vscode.window.activeTextEditor?.document.uri;
    }

    refresh(): void {
        this._onDidChangeCodeLenses.fire();
    }
}