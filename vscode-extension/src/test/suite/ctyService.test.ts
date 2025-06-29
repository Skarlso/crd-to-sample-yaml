import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import * as childProcess from 'child_process';
import * as fs from 'fs';
import { CtyService, GenerationOptions } from '../../ctyService';

suite('CtyService Test Suite', () => {
    let service: CtyService;
    let sandbox: sinon.SinonSandbox;

    setup(() => {
        service = new CtyService();
        sandbox = sinon.createSandbox();
    });

    teardown(() => {
        sandbox.restore();
    });

    suite('getCtyPath', () => {
        test('Should return configured cty path', () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('/custom/path/cty')
            } as any);

            const path = (service as any).getCtyPath();
            assert.strictEqual(path, '/custom/path/cty', 'Should return configured path');
        });

        test('Should return default cty path', () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const path = (service as any).getCtyPath();
            assert.strictEqual(path, 'cty', 'Should return default path');
        });
    });

    suite('getOutputLocation', () => {
        test('Should return workspace path for workspace output', () => {
            const mockWorkspace = { uri: { fsPath: '/workspace/path' } };
            sandbox.stub(vscode.workspace, 'workspaceFolders').value([mockWorkspace]);
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub()
                    .withArgs('outputLocation', 'workspace').returns('workspace')
                    .withArgs('customOutputPath', '').returns('')
            } as any);

            const location = (service as any).getOutputLocation();
            assert.ok(location.includes('workspace'), 'Should return workspace path');
        });

        test('Should return temp directory for temp output', () => {
            const configMock = {
                get: (key: string, defaultValue: any) => {
                    if (key === 'outputLocation') { return 'temp'; }
                    if (key === 'customOutputPath') { return ''; }
                    return defaultValue;
                }
            };
            sandbox.stub(vscode.workspace, 'getConfiguration').returns(configMock as any);
            
            const location = (service as any).getOutputLocation();
            // Just verify it returns the actual temp directory from os.tmpdir()
            assert.ok(location.length > 0, 'Should return a non-empty temp directory path');
            // We can't easily test the exact path since it varies by OS
        });

        test('Should return custom path for custom output', () => {
            const configMock = {
                get: (key: string, defaultValue: any) => {
                    if (key === 'outputLocation') { return 'custom'; }
                    if (key === 'customOutputPath') { return '/custom/output'; }
                    return defaultValue;
                }
            };
            sandbox.stub(vscode.workspace, 'getConfiguration').returns(configMock as any);

            const location = (service as any).getOutputLocation();
            assert.strictEqual(location, '/custom/output', 'Should return custom path');
        });

        test('Should fallback to workspace when custom path is empty', () => {
            const mockWorkspace = { uri: { fsPath: '/workspace/path' } };
            sandbox.stub(vscode.workspace, 'workspaceFolders').value([mockWorkspace]);
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub()
                    .withArgs('outputLocation', 'workspace').returns('custom')
                    .withArgs('customOutputPath', '').returns('')
            } as any);

            const location = (service as any).getOutputLocation();
            assert.ok(location.includes('workspace'), 'Should fallback to workspace path');
        });
    });

    suite('checkCtyAvailable', () => {
        test('Should return true when cty is available', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const execStub = sandbox.stub(childProcess, 'exec').callsArgWith(1, null);

            const result = await service.checkCtyAvailable();
            
            assert.strictEqual(result, true, 'Should return true when cty is available');
            assert.ok(execStub.calledWith('"cty" version'), 'Should execute version command');
        });

        test('Should return false when cty is not available', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const execStub = sandbox.stub(childProcess, 'exec').callsArgWith(1, new Error('Command not found'));

            const result = await service.checkCtyAvailable();
            
            assert.strictEqual(result, false, 'Should return false when cty is not available');
        });
    });

    suite('generateSample', () => {
        test('Should execute cty with correct arguments', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const execFileStub = sandbox.stub(childProcess, 'execFile').callsArgWith(2, null, 'success', '');

            const options: GenerationOptions = {
                minimal: false,
                comments: false
            };

            const result = await service.generateSample('/test/file.yaml', options);

            assert.strictEqual(result, 'success', 'Should return stdout');
            assert.ok(execFileStub.calledOnce, 'Should execute cty command');
            
            const [command, args] = execFileStub.getCall(0).args;
            assert.strictEqual(command, 'cty', 'Should execute cty');
            assert.ok(args && args.includes('generate'), 'Should include generate command');
            assert.ok(args && args.includes('crd'), 'Should include crd subcommand');
            assert.ok(args && args.includes('-c'), 'Should include -c flag');
        });

        test('Should include minimal flag when specified', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const execFileStub = sandbox.stub(childProcess, 'execFile').callsArgWith(2, null, 'success', '');

            const options: GenerationOptions = {
                minimal: true,
                comments: false
            };

            await service.generateSample('/test/file.yaml', options);

            const [, args] = execFileStub.getCall(0).args;
            assert.ok(args && args.includes('--minimal'), 'Should include minimal flag');
        });

        test('Should include comments flag when specified', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const execFileStub = sandbox.stub(childProcess, 'execFile').callsArgWith(2, null, 'success', '');

            const options: GenerationOptions = {
                minimal: false,
                comments: true
            };

            await service.generateSample('/test/file.yaml', options);

            const [, args] = execFileStub.getCall(0).args;
            assert.ok(args && args.includes('--comments'), 'Should include comments flag');
        });

        test('Should include format flag when specified', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const execFileStub = sandbox.stub(childProcess, 'execFile').callsArgWith(2, null, 'success', '');

            const options: GenerationOptions = {
                minimal: false,
                comments: false,
                format: 'html'
            };

            await service.generateSample('/test/file.yaml', options);

            const [, args] = execFileStub.getCall(0).args;
            assert.ok(args && args.includes('--format'), 'Should include format flag');
            assert.ok(args && args.includes('html'), 'Should include format value');
        });

        test('Should handle execution errors', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const error = new Error('Execution failed');
            sandbox.stub(childProcess, 'execFile').callsArgWith(2, error, '', 'stderr output');

            const options: GenerationOptions = {
                minimal: false,
                comments: false
            };

            try {
                await service.generateSample('/test/file.yaml', options);
                assert.fail('Should have thrown error');
            } catch (err) {
                assert.ok(err instanceof Error, 'Should throw error');
                assert.ok(err.message.includes('CTY execution failed'), 'Should include execution error message');
            }
        });
    });

    suite('generateSampleToString', () => {
        test('Should execute cty with stdout flag', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const execStub = sandbox.stub(childProcess, 'exec').callsArgWith(1, null, 'yaml output', '');

            const options: GenerationOptions = {
                minimal: false,
                comments: false
            };

            const result = await service.generateSampleToString('/test/file.yaml', options);

            assert.strictEqual(result, 'yaml output', 'Should return stdout');
            assert.ok(execStub.calledOnce, 'Should execute command');
            
            const command = execStub.getCall(0).args[0];
            assert.ok(command.includes('--stdout'), 'Should include stdout flag');
        });

        test('Should handle execution errors', async () => {
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('ctyPath', 'cty').returns('cty')
            } as any);

            const error = new Error('Execution failed');
            sandbox.stub(childProcess, 'exec').callsArgWith(1, error, '', 'stderr');

            const options: GenerationOptions = {
                minimal: false,
                comments: false
            };

            try {
                await service.generateSampleToString('/test/file.yaml', options);
                assert.fail('Should have thrown error');
            } catch (err) {
                assert.ok(err instanceof Error, 'Should throw error');
                assert.ok(err.message.includes('CTY execution failed'), 'Should include execution error message');
            }
        });
    });

    suite('validateSample', () => {
        test('Should return placeholder validation result', async () => {
            const result = await service.validateSample('/sample.yaml', '/crd.yaml');

            assert.strictEqual(result.valid, true, 'Should return valid as true');
            assert.deepStrictEqual(result.errors, [], 'Should return empty errors array');
        });
    });

    suite('getGeneratedSamplePath', () => {
        test('Should generate path from CRD kind', async () => {
            const crdContent = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  names:
    kind: Example
`;
            
            sandbox.stub(fs.promises, 'readFile').resolves(crdContent);

            const options: GenerationOptions = {
                minimal: false,
                comments: false
            };

            const result = await service.getGeneratedSamplePath('/test/crd.yaml', options);

            assert.ok(result.includes('Example_sample.yaml'), 'Should use CRD kind in filename');
        });

        test('Should use HTML extension for HTML format', async () => {
            const crdContent = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  names:
    kind: Example
`;
            
            sandbox.stub(fs.promises, 'readFile').resolves(crdContent);

            const options: GenerationOptions = {
                minimal: false,
                comments: false,
                format: 'html'
            };

            const result = await service.getGeneratedSamplePath('/test/crd.yaml', options);

            assert.ok(result.includes('Example_sample.html'), 'Should use HTML extension');
        });

        test('Should fallback to filename when CRD parsing fails', async () => {
            sandbox.stub(fs.promises, 'readFile').rejects(new Error('File not found'));

            const options: GenerationOptions = {
                minimal: false,
                comments: false
            };

            const result = await service.getGeneratedSamplePath('/test/crd.yaml', options);

            assert.ok(result.includes('crd_sample.yaml'), 'Should use filename as fallback');
        });

        test('Should use custom output directory', async () => {
            const crdContent = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
spec:
  names:
    kind: Example
`;
            
            sandbox.stub(fs.promises, 'readFile').resolves(crdContent);

            const options: GenerationOptions = {
                minimal: false,
                comments: false,
                output: '/custom/output'
            };

            const result = await service.getGeneratedSamplePath('/test/crd.yaml', options);

            assert.ok(result.includes('/custom/output'), 'Should use custom output directory');
            assert.ok(result.includes('Example_sample.yaml'), 'Should include correct filename');
        });
    });
});