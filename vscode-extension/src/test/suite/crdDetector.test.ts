import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { CRDDetector } from '../../crdDetector';

suite('CRDDetector Test Suite', () => {
    let detector: CRDDetector;
    let sandbox: sinon.SinonSandbox;

    setup(() => {
        detector = new CRDDetector();
        sandbox = sinon.createSandbox();
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
            lineAt: sandbox.stub(),
            offsetAt: sandbox.stub(),
            positionAt: sandbox.stub(),
            getWordRangeAtPosition: sandbox.stub(),
            validateRange: sandbox.stub(),
            validatePosition: sandbox.stub()
        } as any;
    };

    suite('detectCRD', () => {
        test('Should detect valid CRD with CustomResourceDefinition kind', async () => {
            const crdContent = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  versions:
  - name: v1
    served: true
    storage: true
  scope: Namespaced
  names:
    plural: examples
    singular: example
    kind: Example
`;
            const document = createMockDocument(crdContent);
            const result = await detector.detectCRD(document);

            assert.notStrictEqual(result, null, 'Should detect CRD');
            assert.strictEqual(result!.isCRD, true, 'Should be marked as CRD');
            assert.strictEqual(result!.kind, 'Example', 'Should extract correct kind');
            assert.strictEqual(result!.group, 'example.com', 'Should extract correct group');
            assert.strictEqual(result!.version, 'v1', 'Should extract correct version');
            assert.strictEqual(result!.name, 'examples.example.com', 'Should extract correct name');
            assert.strictEqual(result!.hasMultipleVersions, false, 'Should detect single version');
            assert.deepStrictEqual(result!.versions, ['v1'], 'Should extract versions');
        });

        test('Should detect CRD with multiple versions', async () => {
            const crdContent = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  versions:
  - name: v1
    served: true
    storage: true
  - name: v2
    served: true
    storage: false
  scope: Namespaced
  names:
    plural: examples
    singular: example
    kind: Example
`;
            const document = createMockDocument(crdContent);
            const result = await detector.detectCRD(document);

            assert.notStrictEqual(result, null, 'Should detect CRD');
            assert.strictEqual(result!.hasMultipleVersions, true, 'Should detect multiple versions');
            assert.deepStrictEqual(result!.versions, ['v1', 'v2'], 'Should extract all versions');
        });

        test('Should detect CRD with legacy single version format', async () => {
            const crdContent = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: examples.example.com
spec:
  group: example.com
  version: v1
  scope: Namespaced
  names:
    plural: examples
    singular: example
    kind: Example
`;
            const document = createMockDocument(crdContent);
            const result = await detector.detectCRD(document);

            assert.notStrictEqual(result, null, 'Should detect CRD');
            assert.strictEqual(result!.isCRD, true, 'Should be marked as CRD');
            assert.strictEqual(result!.version, 'v1', 'Should extract version from legacy format');
            assert.deepStrictEqual(result!.versions, ['v1'], 'Should extract versions');
        });

        test('Should return null for invalid YAML', async () => {
            const invalidContent = `invalid: yaml: content:`;
            const document = createMockDocument(invalidContent);
            const result = await detector.detectCRD(document);

            assert.strictEqual(result, null, 'Should return null for invalid YAML');
        });

        test('Should return null for non-CRD YAML', async () => {
            const nonCrdContent = `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
spec:
  containers:
  - name: test
    image: nginx
`;
            const document = createMockDocument(nonCrdContent);
            const result = await detector.detectCRD(document);

            assert.strictEqual(result, null, 'Should return null for non-CRD');
        });

        test('Should return null for empty document', async () => {
            const document = createMockDocument('');
            const result = await detector.detectCRD(document);

            assert.strictEqual(result, null, 'Should return null for empty document');
        });
    });

    suite('isCRDDocument', () => {
        test('Should identify CRD by kind', () => {
            const parsed = { kind: 'CustomResourceDefinition' };
            const result = (detector as any).isCRDDocument(parsed);
            assert.strictEqual(result, true, 'Should identify CRD by kind');
        });

        test('Should identify CRD by apiVersion', () => {
            const parsed = { apiVersion: 'apiextensions.k8s.io/v1' };
            const result = (detector as any).isCRDDocument(parsed);
            assert.strictEqual(result, true, 'Should identify CRD by apiVersion');
        });

        test('Should identify CRD by structure', () => {
            const parsed = {
                spec: {
                    names: { kind: 'Example' },
                    group: 'example.com',
                    versions: [{ name: 'v1' }]
                }
            };
            const result = (detector as any).isCRDDocument(parsed);
            assert.strictEqual(result, true, 'Should identify CRD by structure');
        });

        test('Should not identify non-CRD', () => {
            const parsed = { kind: 'Pod', apiVersion: 'v1' };
            const result = (detector as any).isCRDDocument(parsed);
            assert.strictEqual(result, false, 'Should not identify non-CRD');
        });
    });

    suite('detectCRDsInWorkspace', () => {
        test('Should find CRD files in workspace', async () => {
            const mockFiles = [
                vscode.Uri.file('/test1.yaml'),
                vscode.Uri.file('/test2.yaml')
            ];
            
            const workspaceFindFilesStub = sandbox.stub(vscode.workspace, 'findFiles').resolves(mockFiles);
            const openTextDocumentStub = sandbox.stub(vscode.workspace, 'openTextDocument');
            
            const crdDocument = createMockDocument(`
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: test.example.com
spec:
  group: example.com
  versions: [{ name: v1 }]
  names: { kind: Test }
`);
            const nonCrdDocument = createMockDocument(`
apiVersion: v1
kind: Pod
metadata:
  name: test
`);

            openTextDocumentStub.onCall(0).resolves(crdDocument);
            openTextDocumentStub.onCall(1).resolves(nonCrdDocument);

            const result = await detector.detectCRDsInWorkspace();

            assert.strictEqual(workspaceFindFilesStub.callCount, 1, 'Should search for YAML files');
            assert.strictEqual(result.length, 1, 'Should return one CRD file');
            assert.strictEqual(result[0].fsPath, '/test1.yaml', 'Should return correct CRD file');
        });

        test('Should handle file read errors gracefully', async () => {
            const mockFiles = [vscode.Uri.file('/test.yaml')];
            
            sandbox.stub(vscode.workspace, 'findFiles').resolves(mockFiles);
            sandbox.stub(vscode.workspace, 'openTextDocument').rejects(new Error('File not found'));

            const result = await detector.detectCRDsInWorkspace();

            assert.strictEqual(result.length, 0, 'Should return empty array when files cannot be read');
        });
    });

    suite('getCRDLineNumbers', () => {
        test('Should find spec and versions line numbers', () => {
            const content = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: test
spec:
  group: example.com
  versions:
  - name: v1`;
            
            const document = createMockDocument(content);
            const result = detector.getCRDLineNumbers(document);

            assert.strictEqual(result.specLine, 4, 'Should find spec line');
            assert.strictEqual(result.versionsLine, 6, 'Should find versions line');
        });

        test('Should handle missing spec/versions', () => {
            const content = `apiVersion: v1
kind: Pod`;
            
            const document = createMockDocument(content);
            const result = detector.getCRDLineNumbers(document);

            assert.strictEqual(result.specLine, -1, 'Should return -1 for missing spec');
            assert.strictEqual(result.versionsLine, -1, 'Should return -1 for missing versions');
        });
    });

    suite('isValidCRDFile', () => {
        test('Should validate YAML file extensions', () => {
            const yamlUri = vscode.Uri.file('/test.yaml');
            const ymlUri = vscode.Uri.file('/test.yml');
            const txtUri = vscode.Uri.file('/test.txt');

            assert.strictEqual(detector.isValidCRDFile(yamlUri), true, 'Should accept .yaml files');
            assert.strictEqual(detector.isValidCRDFile(ymlUri), true, 'Should accept .yml files');
            assert.strictEqual(detector.isValidCRDFile(txtUri), false, 'Should reject non-YAML files');
        });
    });
});