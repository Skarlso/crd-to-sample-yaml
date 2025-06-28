import {workspace} from "vscode";
import {exec, execFile} from "node:child_process";
import * as path from 'path';
import {promises} from "node:fs";

export interface GenerationOptions {
    minimal: boolean;
    comments: boolean;
    format?: 'yaml' | 'html';
    output?: string;
}

export class CtyService {
    private getCtyPath(): string {
        const config = workspace.getConfiguration('crdToSampleYaml');
        return config.get<string>('ctyPath', 'cty');
    }

    private getOutputLocation(): string {
        const config = workspace.getConfiguration('crdToSampleYaml');
        const outputLocation = config.get<string>('outputLocation', 'workspace');
        const customPath = config.get<string>('customOutputPath', '');

        let workspacePath = '';
        if (workspace.workspaceFolders?.length) {
            workspacePath = workspace.workspaceFolders[0].uri.fsPath;
            workspacePath = path.normalize(workspacePath);
        }

        switch (outputLocation) {
            case 'temp':
                return require('os').tmpdir();
            case 'custom':
                return customPath || workspacePath || '';
            case 'workspace':
                return workspacePath || '';
            default:
                return workspacePath || '';
        }
    }

    async checkCtyAvailable(): Promise<boolean> {
        return new Promise((resolve) => {
            const ctyPath = this.getCtyPath();
            exec(`"${ctyPath}" version`, (error) => {
                resolve(!error);
            });
        });
    }

    async generateSample(filePath: string, options: GenerationOptions): Promise<string> {
        const ctyPath = this.getCtyPath();
        const outputDir = this.getOutputLocation();
        
        const args = ['generate', 'crd', '-c', path.resolve(filePath)];
        
        if (options.minimal) {
            args.push('--minimal');
        }
        
        if (options.comments) {
            args.push('--comments');
        }
        
        if (options.format && options.format !== 'yaml') {
            args.push('--format', options.format);
        }

        const targetOutputDir = options.output || outputDir || path.dirname(filePath);
        const absoluteOutputDir = path.resolve(targetOutputDir);
        args.push('--output', absoluteOutputDir);

        return new Promise((resolve, reject) => {
            const command = `"${ctyPath}" ${args.join(' ')}`;

            execFile(ctyPath, args, { cwd: absoluteOutputDir }, (error, stdout, stderr) => {
                if (error) {
                    reject(new Error(`CTY execution failed: ${error.message}\nStderr: ${stderr}`));
                    return;
                }
                
                resolve(stdout);
            });
        });
    }

    async generateSampleToString(filePath: string, options: GenerationOptions): Promise<string> {
        const ctyPath = this.getCtyPath();
        
        const args = ['generate', 'crd', '-c', filePath, '--stdout'];
        
        if (options.minimal) {
            args.push('--minimal');
        }
        
        if (options.comments) {
            args.push('--comments');
        }

        return new Promise((resolve, reject) => {
            const command = `"${ctyPath}" ${args.join(' ')}`;
            
            exec(command, (error, stdout, stderr) => {
                if (error) {
                    reject(new Error(`CTY execution failed: ${error.message}\nStderr: ${stderr}`));
                    return;
                }
                
                resolve(stdout);
            });
        });
    }

    async validateSample(samplePath: string, crdPath: string): Promise<{ valid: boolean; errors: string[] }> {
        return new Promise((resolve) => {
            // TODO: finish this
            resolve({ valid: true, errors: [] });
        });
    }

    async getGeneratedSamplePath(crdPath: string, options: GenerationOptions): Promise<string> {
        const outputDir = this.getOutputLocation();
        const targetOutputDir = options.output || outputDir || path.dirname(crdPath);
        const absoluteOutputDir = path.resolve(targetOutputDir);
        
        try {
            const crdContent = await promises.readFile(crdPath, 'utf8');
            const yaml = require('js-yaml');
            const parsed = yaml.load(crdContent) as any;
            const kind = parsed.spec?.names?.kind || path.basename(crdPath, path.extname(crdPath));
            const extension = options.format === 'html' ? 'html' : 'yaml';
            
            return path.join(absoluteOutputDir, `${kind}_sample.${extension}`);
        } catch (error) {
            const crdName = path.basename(crdPath, path.extname(crdPath));
            const extension = options.format === 'html' ? 'html' : 'yaml';
            return path.join(absoluteOutputDir, `${crdName}_sample.${extension}`);
        }
    }
}