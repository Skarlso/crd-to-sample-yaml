import * as assert from 'assert';
import * as vscode from 'vscode';
import * as sinon from 'sinon';
import { activate, deactivate } from '../../extension';

suite('Extension Test Suite', () => {
    vscode.window.showInformationMessage('Start all tests.');

    let context: vscode.ExtensionContext;
    let sandbox: sinon.SinonSandbox;

    setup(() => {
        sandbox = sinon.createSandbox();
        context = {
            subscriptions: [],
            workspaceState: {
                get: sandbox.stub() as any,
                update: sandbox.stub() as any,
                keys: sandbox.stub() as any
            },
            globalState: {
                get: sandbox.stub().returns(false) as any,
                update: sandbox.stub() as any,
                setKeysForSync: sandbox.stub() as any,
                keys: sandbox.stub() as any
            },
            secrets: {} as any,
            extensionUri: vscode.Uri.file('/test'),
            extensionPath: '/test',
            environmentVariableCollection: {} as any,
            asAbsolutePath: sandbox.stub(),
            storageUri: vscode.Uri.file('/test'),
            storagePath: '/test',
            globalStorageUri: vscode.Uri.file('/test'),
            globalStoragePath: '/test',
            logUri: vscode.Uri.file('/test'),
            logPath: '/test',
            extensionMode: vscode.ExtensionMode.Test,
            extension: {} as any,
            languageModelAccessInformation: {} as any
        };
    });

    teardown(() => {
        sandbox.restore();
    });

    test('Extension should activate without errors', async () => {
        const commandsRegisterStub = sandbox.stub(vscode.commands, 'registerCommand');
        const languagesRegisterCodeLensStub = sandbox.stub(vscode.languages, 'registerCodeLensProvider');
        const languagesRegisterHoverStub = sandbox.stub(vscode.languages, 'registerHoverProvider');
        const workspaceCreateFileSystemWatcherStub = sandbox.stub(vscode.workspace, 'createFileSystemWatcher');
        const windowShowInformationMessageStub = sandbox.stub(vscode.window, 'showInformationMessage');

        const mockWatcher = {
            onDidChange: sandbox.stub()
        };
        workspaceCreateFileSystemWatcherStub.returns(mockWatcher as any);
        windowShowInformationMessageStub.returns(Promise.resolve('Got it!') as any);

        activate(context);

        assert.strictEqual(commandsRegisterStub.callCount, 4, 'Should register 4 commands');
        assert.strictEqual(languagesRegisterCodeLensStub.callCount, 1, 'Should register code lens provider');
        assert.strictEqual(languagesRegisterHoverStub.callCount, 1, 'Should register hover provider');
        assert.strictEqual(workspaceCreateFileSystemWatcherStub.callCount, 1, 'Should create file system watcher');
        assert.strictEqual(context.subscriptions.length, 7, 'Should have 7 subscriptions');
    });

    test('Extension should register correct commands', async () => {
        const commandsRegisterStub = sandbox.stub(vscode.commands, 'registerCommand');
        sandbox.stub(vscode.languages, 'registerCodeLensProvider');
        sandbox.stub(vscode.languages, 'registerHoverProvider');
        sandbox.stub(vscode.workspace, 'createFileSystemWatcher').returns({
            onDidChange: sandbox.stub()
        } as any);
        sandbox.stub(vscode.window, 'showInformationMessage').returns(Promise.resolve('Got it!') as any);

        activate(context);

        const expectedCommands = [
            'crdToSampleYaml.generateSample',
            'crdToSampleYaml.generateMinimalSample',
            'crdToSampleYaml.generateSampleWithComments',
            'crdToSampleYaml.validateSample'
        ];

        expectedCommands.forEach((command, index) => {
            assert.strictEqual(
                commandsRegisterStub.getCall(index).args[0], 
                command, 
                `Should register command: ${command}`
            );
        });
    });

    test('Extension should show welcome message on first activation', async () => {
        const windowShowInformationMessageStub = sandbox.stub(vscode.window, 'showInformationMessage');
        sandbox.stub(vscode.commands, 'registerCommand');
        sandbox.stub(vscode.languages, 'registerCodeLensProvider');
        sandbox.stub(vscode.languages, 'registerHoverProvider');
        sandbox.stub(vscode.workspace, 'createFileSystemWatcher').returns({
            onDidChange: sandbox.stub()
        } as any);

        const thenStub = sandbox.stub().resolves();
        windowShowInformationMessageStub.returns({ then: thenStub } as any);

        activate(context);

        assert.strictEqual(windowShowInformationMessageStub.callCount, 1, 'Should show welcome message');
        assert.strictEqual(
            windowShowInformationMessageStub.getCall(0).args[0],
            'CRD to Sample YAML extension activated! Right-click on CRD files to generate samples.',
            'Should show correct welcome message'
        );
    });

    test('Extension should not show welcome message if already shown', async () => {
        const windowShowInformationMessageStub = sandbox.stub(vscode.window, 'showInformationMessage');
        sandbox.stub(vscode.commands, 'registerCommand');
        sandbox.stub(vscode.languages, 'registerCodeLensProvider');
        sandbox.stub(vscode.languages, 'registerHoverProvider');  
        sandbox.stub(vscode.workspace, 'createFileSystemWatcher').returns({
            onDidChange: sandbox.stub()
        } as any);

        (context.globalState as any).get = sandbox.stub().returns(true);

        activate(context);

        assert.strictEqual(windowShowInformationMessageStub.callCount, 0, 'Should not show welcome message');
    });

    test('Deactivate should work without errors', () => {
        assert.doesNotThrow(() => {
            deactivate();
        }, 'Deactivate should not throw');
    });
});