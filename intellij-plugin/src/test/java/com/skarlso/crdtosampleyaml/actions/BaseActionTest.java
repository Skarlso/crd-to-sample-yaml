package com.skarlso.crdtosampleyaml.actions;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.CommonDataKeys;
import com.intellij.openapi.actionSystem.Presentation;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.testFramework.fixtures.BasePlatformTestCase;
import com.skarlso.crdtosampleyaml.services.CrdDetectorService;
import org.junit.Test;
import org.mockito.Mock;
import org.mockito.MockitoAnnotations;

import static org.mockito.Mockito.*;

public class BaseActionTest extends BasePlatformTestCase {
    
    private TestableBaseAction baseAction;
    
    @Mock
    private AnActionEvent mockEvent;
    
    @Mock
    private Presentation mockPresentation;
    
    private static class TestableBaseAction extends BaseAction {
        @Override
        public void actionPerformed(AnActionEvent e) {
            // Test implementation - no-op
        }
    }
    
    @Override
    protected void setUp() throws Exception {
        super.setUp();
        MockitoAnnotations.openMocks(this);
        baseAction = new TestableBaseAction();
    }
    
    @Test
    public void testUpdate_ValidCrdFile_EnablesAction() throws Exception {
        VirtualFile crdFile = myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            spec:
              group: example.com
              versions:
              - name: v1
                served: true
                storage: true
                schema:
                  openAPIV3Schema:
                    type: object
              scope: Namespaced
              names:
                plural: myresources
                singular: myresource
                kind: MyResource
            """);
        
        when(mockEvent.getProject()).thenReturn(getProject());
        when(mockEvent.getData(CommonDataKeys.VIRTUAL_FILE)).thenReturn(crdFile);
        when(mockEvent.getPresentation()).thenReturn(mockPresentation);
        
        baseAction.update(mockEvent);
        
        verify(mockPresentation).setEnabledAndVisible(true);
    }
    
    @Test
    public void testUpdate_NonCrdFile_DisablesAction() throws Exception {
        VirtualFile nonCrdFile = myFixture.createFile("test-pod.yaml", """
            apiVersion: v1
            kind: Pod
            metadata:
              name: test-pod
            spec:
              containers:
              - name: test
                image: nginx
            """);
        
        when(mockEvent.getProject()).thenReturn(getProject());
        when(mockEvent.getData(CommonDataKeys.VIRTUAL_FILE)).thenReturn(nonCrdFile);
        when(mockEvent.getPresentation()).thenReturn(mockPresentation);
        
        baseAction.update(mockEvent);
        
        verify(mockPresentation).setEnabledAndVisible(false);
    }
    
    @Test
    public void testUpdate_NoProject_DisablesAction() {
        when(mockEvent.getProject()).thenReturn(null);
        when(mockEvent.getData(CommonDataKeys.VIRTUAL_FILE)).thenReturn(mock(VirtualFile.class));
        when(mockEvent.getPresentation()).thenReturn(mockPresentation);
        
        baseAction.update(mockEvent);
        
        verify(mockPresentation).setEnabledAndVisible(false);
    }
    
    @Test
    public void testUpdate_NoFile_DisablesAction() {
        when(mockEvent.getProject()).thenReturn(getProject());
        when(mockEvent.getData(CommonDataKeys.VIRTUAL_FILE)).thenReturn(null);
        when(mockEvent.getPresentation()).thenReturn(mockPresentation);
        
        baseAction.update(mockEvent);
        
        verify(mockPresentation).setEnabledAndVisible(false);
    }
    
    @Test
    public void testGetSelectedFile_ReturnsFileFromEvent() {
        VirtualFile expectedFile = mock(VirtualFile.class);
        when(mockEvent.getData(CommonDataKeys.VIRTUAL_FILE)).thenReturn(expectedFile);
        
        VirtualFile result = baseAction.getSelectedFile(mockEvent);
        
        assertEquals("Should return file from event", expectedFile, result);
    }
    
    @Test
    public void testGetProject_ReturnsProjectFromEvent() {
        Project expectedProject = getProject();
        when(mockEvent.getProject()).thenReturn(expectedProject);
        
        Project result = baseAction.getProject(mockEvent);
        
        assertEquals("Should return project from event", expectedProject, result);
    }
}