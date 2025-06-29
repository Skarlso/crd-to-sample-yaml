import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { CRDProvider } from '../../crdProvider';
import { CtyService } from '../../ctyService';
import { CRDDetector, CRDInfo } from '../../crdDetector';

suite('CRDProvider Test Suite', () => {
    let provider: CRDProvider;
    let mockCtyService: sinon.SinonStubbedInstance<CtyService>;
    let mockCrdDetector: sinon.SinonStubbedInstance<CRDDetector>;
    let sandbox: sinon.SinonSandbox;

    setup(() => {
        sandbox = sinon.createSandbox();
        mockCtyService = sandbox.createStubInstance(CtyService);
        mockCrdDetector = sandbox.createStubInstance(CRDDetector);
        provider = new CRDProvider(mockCtyService, mockCrdDetector);
    });

    teardown(() => {
        sandbox.restore();
    });

    const createMockDocument = (content: string): vscode.TextDocument => {
        return {
            getText: () => content,
            uri: vscode.Uri.file('/test.yaml'),
            fileName: '/test.yaml',
            languageId: 'yaml',
            version: 1,
            isDirty: false,
            isClosed: false,
            save: sandbox.stub(),
            eol: vscode.EndOfLine.LF,
            lineCount: content.split('\n').length,
            lineAt: (line: number) => ({
                text: content.split('\n')[line] || '',
                lineNumber: line,
                range: new vscode.Range(line, 0, line, 0),
                rangeIncludingLineBreak: new vscode.Range(line, 0, line, 0),
                firstNonWhitespaceCharacterIndex: 0,
                isEmptyOrWhitespace: false
            }),
            offsetAt: sandbox.stub(),
            positionAt: sandbox.stub(),
            getWordRangeAtPosition: sandbox.stub(),
            validateRange: sandbox.stub(),
            validatePosition: sandbox.stub()
        } as any;
    };

    const mockCRDInfo: CRDInfo = {
        kind: 'Example',
        group: 'example.com',
        version: 'v1',
        name: 'examples.example.com',
        isCRD: true,
        hasMultipleVersions: false,
        versions: ['v1']
    };

    suite('provideCodeLenses', () => {
        test('Should provide code lenses for CRD files', async () => {
            const document = createMockDocument('test content');
            
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('autoDetectCRDs', true).returns(true)
            } as any);

            mockCrdDetector.detectCRD.resolves(mockCRDInfo);
            mockCrdDetector.getCRDLineNumbers.returns({ specLine: 5, versionsLine: 7 });

            const result = await provider.provideCodeLenses(document);

            assert.strictEqual(result.length, 3, 'Should provide 3 code lenses');
            assert.strictEqual(result[0].command!.title, '🔧 Generate Sample');
            assert.strictEqual(result[1].command!.title, '📝 Generate Minimal');
            assert.strictEqual(result[2].command!.title, '💬 Generate with Comments');
        });

        test('Should return empty array when auto-detect is disabled', async () => {
            const document = createMockDocument('test content');
            
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('autoDetectCRDs', true).returns(false)
            } as any);

            const result = await provider.provideCodeLenses(document);

            assert.strictEqual(result.length, 0, 'Should return empty array when auto-detect disabled');
        });

        test('Should return empty array for non-CRD files', async () => {
            const document = createMockDocument('test content');
            
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('autoDetectCRDs', true).returns(true)
            } as any);

            mockCrdDetector.detectCRD.resolves(null);

            const result = await provider.provideCodeLenses(document);

            assert.strictEqual(result.length, 0, 'Should return empty array for non-CRD files');
        });

        test('Should return empty array when spec line not found', async () => {
            const document = createMockDocument('test content');
            
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub().withArgs('autoDetectCRDs', true).returns(true)
            } as any);

            mockCrdDetector.detectCRD.resolves(mockCRDInfo);
            mockCrdDetector.getCRDLineNumbers.returns({ specLine: -1, versionsLine: -1 });

            const result = await provider.provideCodeLenses(document);

            assert.strictEqual(result.length, 0, 'Should return empty array when spec line not found');
        });
    });

    suite('provideHover', () => {
        test('Should provide hover information for CRD kind line', async () => {
            const document = createMockDocument('kind: CustomResourceDefinition');
            const position = new vscode.Position(0, 5);

            mockCrdDetector.detectCRD.resolves(mockCRDInfo);

            const result = await provider.provideHover(document, position);

            assert.notStrictEqual(result, undefined, 'Should provide hover');
            assert.ok(result!.contents[0] instanceof vscode.MarkdownString, 'Should return MarkdownString');
        });

        test('Should return undefined for non-CRD files', async () => {
            const document = createMockDocument('kind: Pod');
            const position = new vscode.Position(0, 5);

            mockCrdDetector.detectCRD.resolves(null);

            const result = await provider.provideHover(document, position);

            assert.strictEqual(result, undefined, 'Should return undefined for non-CRD files');
        });

        test('Should return undefined for non-kind lines', async () => {
            const document = createMockDocument('apiVersion: v1');
            const position = new vscode.Position(0, 5);

            mockCrdDetector.detectCRD.resolves(mockCRDInfo);

            const result = await provider.provideHover(document, position);

            assert.strictEqual(result, undefined, 'Should return undefined for non-kind lines');
        });
    });

    suite('generateSample', () => {
        test('Should generate sample successfully', async () => {
            const testUri = vscode.Uri.file('/test.yaml');
            const document = createMockDocument('test content');

            mockCtyService.checkCtyAvailable.resolves(true);
            mockCrdDetector.isValidCRDFile.returns(true);
            mockCrdDetector.detectCRD.resolves(mockCRDInfo);
            mockCtyService.generateSample.resolves('success');
            mockCtyService.getGeneratedSamplePath.resolves('/output/Example_sample.yaml');

            sandbox.stub(vscode.workspace, 'openTextDocument')
                .onFirstCall().resolves(document as any)
                .onSecondCall().resolves(document as any);

            sandbox.stub(vscode.window, 'showTextDocument');
            sandbox.stub(vscode.window, 'withProgress').callsFake(async (options, task) => {
                return task({ report: sandbox.stub() }, {} as any);
            });

            const fsStub = {
                existsSync: sandbox.stub().returns(true)
            };
            sandbox.stub(require('module'), '_load').withArgs('fs').returns(fsStub);

            await provider.generateSample(testUri);

            assert.ok(mockCtyService.generateSample.calledOnce, 'Should call generateSample');
            assert.ok(mockCtyService.getGeneratedSamplePath.calledOnce, 'Should get generated sample path');
        });

        test('Should show error when cty is not available', async () => {
            const testUri = vscode.Uri.file('/test.yaml');
            
            mockCtyService.checkCtyAvailable.resolves(false);
            
            const showErrorMessageStub = sandbox.stub(vscode.window, 'showErrorMessage').resolves('Configure Path' as any);
            const executeCommandStub = sandbox.stub(vscode.commands, 'executeCommand');

            await provider.generateSample(testUri);

            assert.ok(showErrorMessageStub.calledOnce, 'Should show error message');
            assert.ok(executeCommandStub.calledWith('workbench.action.openSettings', 'crdToSampleYaml.ctyPath'), 'Should open settings');
        });

        test('Should show error for invalid file', async () => {
            const testUri = vscode.Uri.file('/test.txt');
            
            mockCtyService.checkCtyAvailable.resolves(true);
            mockCrdDetector.isValidCRDFile.returns(false);
            
            const showErrorMessageStub = sandbox.stub(vscode.window, 'showErrorMessage');

            await provider.generateSample(testUri);

            assert.ok(showErrorMessageStub.calledWith('Selected file is not a valid YAML file'), 'Should show invalid file error');
        });

        test('Should show error for non-CRD file', async () => {
            const testUri = vscode.Uri.file('/test.yaml');
            const document = createMockDocument('test content');
            
            mockCtyService.checkCtyAvailable.resolves(true);
            mockCrdDetector.isValidCRDFile.returns(true);
            mockCrdDetector.detectCRD.resolves(null);
            
            sandbox.stub(vscode.workspace, 'openTextDocument').resolves(document as any);
            const showErrorMessageStub = sandbox.stub(vscode.window, 'showErrorMessage');

            await provider.generateSample(testUri);

            assert.ok(showErrorMessageStub.calledWith('Selected file does not appear to be a CRD'), 'Should show non-CRD error');
        });

        test('Should handle generation errors', async () => {
            const testUri = vscode.Uri.file('/test.yaml');
            const document = createMockDocument('test content');
            
            mockCtyService.checkCtyAvailable.resolves(true);
            mockCrdDetector.isValidCRDFile.returns(true);
            mockCrdDetector.detectCRD.resolves(mockCRDInfo);
            mockCtyService.generateSample.rejects(new Error('Generation failed'));

            sandbox.stub(vscode.workspace, 'openTextDocument').resolves(document as any);
            sandbox.stub(vscode.window, 'withProgress').callsFake(async (options, task) => {
                return task({ report: sandbox.stub() }, {} as any);
            });
            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub()
                    .withArgs('outputLocation', 'workspace').returns('workspace')
                    .withArgs('customOutputPath', '').returns('')
            } as any);

            const showErrorMessageStub = sandbox.stub(vscode.window, 'showErrorMessage');

            await provider.generateSample(testUri);

            assert.ok(showErrorMessageStub.calledWith('Failed to generate sample: Generation failed'), 'Should show generation error');
        });

        test('Should use active document when no URI provided', async () => {
            const testUri = vscode.Uri.file('/test.yaml');
            const document = createMockDocument('test content');
            
            sandbox.stub(vscode.window, 'activeTextEditor').value({
                document: { uri: testUri }
            });

            mockCtyService.checkCtyAvailable.resolves(true);
            mockCrdDetector.isValidCRDFile.returns(true);
            mockCrdDetector.detectCRD.resolves(mockCRDInfo);
            mockCtyService.generateSample.resolves('success');
            mockCtyService.getGeneratedSamplePath.resolves('/output/Example_sample.yaml');

            sandbox.stub(vscode.workspace, 'openTextDocument')
                .onFirstCall().resolves(document as any)
                .onSecondCall().resolves(document as any);
            
            sandbox.stub(vscode.window, 'showTextDocument');
            sandbox.stub(vscode.window, 'withProgress').callsFake(async (options, task) => {
                return task({ report: sandbox.stub() }, {} as any);
            });

            sandbox.stub(vscode.workspace, 'getConfiguration').returns({
                get: sandbox.stub()
                    .withArgs('outputLocation', 'workspace').returns('workspace')
                    .withArgs('customOutputPath', '').returns('')
                    .withArgs('showNotifications', true).returns(true)
            } as any);

            // Mock fs.existsSync
            const fsStub = {
                existsSync: sandbox.stub().returns(true)
            };
            sandbox.stub(require('module'), '_load').withArgs('fs').returns(fsStub);

            await provider.generateSample();

            assert.ok(mockCtyService.generateSample.calledOnce, 'Should generate sample for active document');
        });
    });

    suite('validateSample', () => {
        test('Should show coming soon message', async () => {
            const testUri = vscode.Uri.file('/test.yaml');
            const showInfoStub = sandbox.stub(vscode.window, 'showInformationMessage');

            await provider.validateSample(testUri);

            assert.ok(showInfoStub.calledWith('Sample validation feature coming soon!'), 'Should show coming soon message');
        });
    });

    suite('refresh', () => {
        test('Should fire code lens change event', () => {
            const eventSpy = sandbox.spy();
            provider.onDidChangeCodeLenses(eventSpy);

            provider.refresh();

            assert.ok(eventSpy.calledOnce, 'Should fire code lens change event');
        });
    });
});