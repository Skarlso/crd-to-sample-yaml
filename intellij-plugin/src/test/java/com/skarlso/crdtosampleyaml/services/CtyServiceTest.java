package com.skarlso.crdtosampleyaml.services;

import com.intellij.execution.ExecutionException;
import com.intellij.execution.configurations.GeneralCommandLine;
import com.intellij.execution.process.ProcessOutput;
import com.intellij.execution.util.ExecUtil;
import com.intellij.notification.Notification;
import com.intellij.notification.NotificationGroup;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.testFramework.fixtures.BasePlatformTestCase;
import com.skarlso.crdtosampleyaml.settings.CtySettings;
import org.junit.Test;
import org.mockito.MockedStatic;
import org.mockito.Mockito;

import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

public class CtyServiceTest extends BasePlatformTestCase {
    
    private CtyService ctyService;
    private CtySettings mockSettings;
    
    @Override
    protected void setUp() throws Exception {
        super.setUp();
        ctyService = new CtyService(getProject());
        mockSettings = mock(CtySettings.class);
    }
    
    @Test
    public void testGenerateSample_SuccessfulGeneration() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
        
        ProcessOutput successOutput = mock(ProcessOutput.class);
        when(successOutput.getExitCode()).thenReturn(0);
        when(successOutput.getStdout()).thenReturn("Sample generated successfully");
        
        try (MockedStatic<ExecUtil> execUtilMock = mockStatic(ExecUtil.class);
             MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            
            when(CtySettings.getInstance()).thenReturn(mockSettings);
            when(mockSettings.getCtyPath()).thenReturn("/usr/local/bin/cty");
            execUtilMock.when(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)))
                       .thenReturn(successOutput);
            
            ctyService.generateSample(crdFile, CtyService.GenerationType.COMPLETE);
            
            execUtilMock.verify(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)), times(1));
        }
    }
    
    @Test
    public void testGenerateSample_CtyNotFound() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
        
        try (MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class);
             MockedStatic<ExecUtil> execUtilMock = mockStatic(ExecUtil.class)) {
            
            when(CtySettings.getInstance()).thenReturn(mockSettings);
            when(mockSettings.getCtyPath()).thenReturn("");
            
            ProcessOutput whichOutput = mock(ProcessOutput.class);
            when(whichOutput.getExitCode()).thenReturn(1);
            execUtilMock.when(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)))
                       .thenReturn(whichOutput);
            
            ctyService.generateSample(crdFile, CtyService.GenerationType.COMPLETE);
            
            execUtilMock.verify(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)), times(1));
        }
    }
    
    @Test
    public void testGenerateSample_ExecutionFailure() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
        
        ProcessOutput failureOutput = mock(ProcessOutput.class);
        when(failureOutput.getExitCode()).thenReturn(1);
        when(failureOutput.getStderr()).thenReturn("Error: Invalid CRD format");
        
        try (MockedStatic<ExecUtil> execUtilMock = mockStatic(ExecUtil.class);
             MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            
            when(CtySettings.getInstance()).thenReturn(mockSettings);
            when(mockSettings.getCtyPath()).thenReturn("/usr/local/bin/cty");
            execUtilMock.when(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)))
                       .thenReturn(failureOutput);
            
            ctyService.generateSample(crdFile, CtyService.GenerationType.COMPLETE);
            
            execUtilMock.verify(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)), times(1));
        }
    }
    
    @Test
    public void testGenerateSample_MinimalType() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
        
        ProcessOutput successOutput = mock(ProcessOutput.class);
        when(successOutput.getExitCode()).thenReturn(0);
        
        try (MockedStatic<ExecUtil> execUtilMock = mockStatic(ExecUtil.class);
             MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            
            when(CtySettings.getInstance()).thenReturn(mockSettings);
            when(mockSettings.getCtyPath()).thenReturn("/usr/local/bin/cty");
            execUtilMock.when(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)))
                       .thenReturn(successOutput);
            
            ctyService.generateSample(crdFile, CtyService.GenerationType.MINIMAL);
            
            execUtilMock.verify(() -> ExecUtil.execAndGetOutput(argThat(cmd -> 
                cmd.getCommandLineString().contains("-l"))), times(1));
        }
    }
    
    @Test
    public void testGenerateSample_WithCommentsType() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
        
        ProcessOutput successOutput = mock(ProcessOutput.class);
        when(successOutput.getExitCode()).thenReturn(0);
        
        try (MockedStatic<ExecUtil> execUtilMock = mockStatic(ExecUtil.class);
             MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            
            when(CtySettings.getInstance()).thenReturn(mockSettings);
            when(mockSettings.getCtyPath()).thenReturn("/usr/local/bin/cty");
            execUtilMock.when(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)))
                       .thenReturn(successOutput);
            
            ctyService.generateSample(crdFile, CtyService.GenerationType.WITH_COMMENTS);
            
            execUtilMock.verify(() -> ExecUtil.execAndGetOutput(argThat(cmd -> 
                cmd.getCommandLineString().contains("-m"))), times(1));
        }
    }
    
    @Test
    public void testValidateSample_SuccessfulValidation() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
        
        VirtualFile sampleFile = myFixture.createFile("sample.yaml", """
            apiVersion: example.com/v1
            kind: MyResource
            metadata:
              name: test-resource
            """);
        
        ProcessOutput successOutput = mock(ProcessOutput.class);
        when(successOutput.getExitCode()).thenReturn(0);
        
        try (MockedStatic<ExecUtil> execUtilMock = mockStatic(ExecUtil.class);
             MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            
            when(CtySettings.getInstance()).thenReturn(mockSettings);
            when(mockSettings.getCtyPath()).thenReturn("/usr/local/bin/cty");
            execUtilMock.when(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)))
                       .thenReturn(successOutput);
            
            ctyService.validateSample(sampleFile, crdFile);
            
            execUtilMock.verify(() -> ExecUtil.execAndGetOutput(argThat(cmd -> 
                cmd.getCommandLineString().contains("validate") &&
                cmd.getCommandLineString().contains("-c") &&
                cmd.getCommandLineString().contains("-s"))), times(1));
        }
    }
    
    @Test
    public void testValidateSample_ValidationFailure() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
        
        VirtualFile sampleFile = myFixture.createFile("invalid-sample.yaml", """
            apiVersion: example.com/v1
            kind: MyResource
            metadata:
              name: test-resource
            spec:
              invalidField: value
            """);
        
        ProcessOutput failureOutput = mock(ProcessOutput.class);
        when(failureOutput.getExitCode()).thenReturn(1);
        when(failureOutput.getStderr()).thenReturn("Validation failed: unknown field 'invalidField'");
        
        try (MockedStatic<ExecUtil> execUtilMock = mockStatic(ExecUtil.class);
             MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            
            when(CtySettings.getInstance()).thenReturn(mockSettings);
            when(mockSettings.getCtyPath()).thenReturn("/usr/local/bin/cty");
            execUtilMock.when(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)))
                       .thenReturn(failureOutput);
            
            ctyService.validateSample(sampleFile, crdFile);
            
            execUtilMock.verify(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)), times(1));
        }
    }
    
    @Test
    public void testValidateSample_ExecutionException() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
        
        VirtualFile sampleFile = myFixture.createFile("sample.yaml", """
            apiVersion: example.com/v1
            kind: MyResource
            metadata:
              name: test-resource
            """);
        
        try (MockedStatic<ExecUtil> execUtilMock = mockStatic(ExecUtil.class);
             MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            
            when(CtySettings.getInstance()).thenReturn(mockSettings);
            when(mockSettings.getCtyPath()).thenReturn("/usr/local/bin/cty");
            execUtilMock.when(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)))
                       .thenThrow(new ExecutionException("Command execution failed"));
            
            ctyService.validateSample(sampleFile, crdFile);
            
            execUtilMock.verify(() -> ExecUtil.execAndGetOutput(any(GeneralCommandLine.class)), times(1));
        }
    }
}