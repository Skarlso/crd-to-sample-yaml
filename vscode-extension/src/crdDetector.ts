import * as vscode from 'vscode';
import * as yaml from 'js-yaml';

export interface CRDInfo {
    kind: string;
    group: string;
    version: string;
    name: string;
    isCRD: boolean;
    hasMultipleVersions: boolean;
    versions: string[];
}

export class CRDDetector {
    async detectCRD(document: vscode.TextDocument): Promise<CRDInfo | null> {
        try {
            const content = document.getText();
            const parsed = yaml.load(content) as any;
            
            if (!parsed || typeof parsed !== 'object') {
                return null;
            }

            // Check if this is a CRD
            const isCRD = this.isCRDDocument(parsed);
            if (!isCRD) {
                return null;
            }

            const kind = parsed.spec?.names?.kind || 'Unknown';
            const group = parsed.spec?.group || 'Unknown';
            const name = parsed.metadata?.name || 'Unknown';
            
            // Extract versions
            const versions: string[] = [];
            let hasMultipleVersions = false;
            
            if (parsed.spec?.versions && Array.isArray(parsed.spec.versions)) {
                hasMultipleVersions = parsed.spec.versions.length > 1;
                versions.push(...parsed.spec.versions.map((v: any) => v.name || v.version));
            } else if (parsed.spec?.version) {
                versions.push(parsed.spec.version);
            }

            const version = versions.length > 0 ? versions[0] : 'Unknown';

            return {
                kind,
                group,
                version,
                name,
                isCRD: true,
                hasMultipleVersions,
                versions
            };
        } catch (error) {
            // Not a valid YAML or not a CRD
            return null;
        }
    }

    private isCRDDocument(parsed: any): boolean {
        // Check for CustomResourceDefinition kind
        if (parsed.kind === 'CustomResourceDefinition') {
            return true;
        }

        // Check for apiVersion containing apiextensions
        if (parsed.apiVersion && 
            typeof parsed.apiVersion === 'string' && 
            parsed.apiVersion.includes('apiextensions')) {
            return true;
        }

        // Check for CRD-like structure
        return !!(parsed.spec &&
            parsed.spec.names &&
            parsed.spec.group &&
            (parsed.spec.versions || parsed.spec.version));
    }

    async detectCRDsInWorkspace(): Promise<vscode.Uri[]> {
        const crdFiles: vscode.Uri[] = [];
        
        // Find all YAML files in a workspace
        const yamlFiles = await vscode.workspace.findFiles('**/*.{yaml,yml}', '**/node_modules/**');
        
        for (const file of yamlFiles) {
            try {
                const document = await vscode.workspace.openTextDocument(file);
                const crdInfo = await this.detectCRD(document);
                
                if (crdInfo?.isCRD) {
                    crdFiles.push(file);
                }
            } catch (error) {
                // Skip files that can't be read or parsed
            }
        }
        
        return crdFiles;
    }

    getCRDLineNumbers(document: vscode.TextDocument): { specLine: number; versionsLine: number } {
        const text = document.getText();
        const lines = text.split('\n');
        
        let specLine = -1;
        let versionsLine = -1;
        
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i].trim();
            
            if (line.startsWith('spec:')) {
                specLine = i;
            }
            
            if (line.startsWith('versions:') || line.startsWith('version:')) {
                versionsLine = i;
            }
        }
        
        return { specLine, versionsLine };
    }

    isValidCRDFile(uri: vscode.Uri): boolean {
        const fileName = uri.fsPath.toLowerCase();
        return fileName.endsWith('.yaml') || fileName.endsWith('.yml');
    }
}