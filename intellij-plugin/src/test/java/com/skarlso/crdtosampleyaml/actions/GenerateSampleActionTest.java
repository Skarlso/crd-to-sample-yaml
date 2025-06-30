package com.skarlso.crdtosampleyaml.actions;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.CommonDataKeys;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.testFramework.fixtures.BasePlatformTestCase;
import com.skarlso.crdtosampleyaml.services.CtyService;
import org.junit.Test;
import org.mockito.Mock;
import org.mockito.MockedConstruction;
import org.mockito.MockitoAnnotations;

import static org.mockito.Mockito.*;

public class GenerateSampleActionTest extends BasePlatformTestCase {
    
    private GenerateSampleAction action;
    
    @Mock
    private AnActionEvent mockEvent;
    
    @Override
    protected void setUp() throws Exception {
        super.setUp();
        MockitoAnnotations.openMocks(this);
        action = new GenerateSampleAction();
    }
    
    @Test
    public void testActionPerformed_ValidCrdFile_CallsService() throws Exception {
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
        
        try (MockedConstruction<CtyService> mockedService = mockConstruction(CtyService.class)) {
            action.actionPerformed(mockEvent);
            
            assertEquals("Should create one CtyService instance", 1, mockedService.constructed().size());
            CtyService serviceInstance = mockedService.constructed().get(0);
            verify(serviceInstance).generateSample(crdFile, CtyService.GenerationType.COMPLETE);
        }
    }
    
    @Test
    public void testActionPerformed_NoProject_DoesNotCallService() {
        VirtualFile crdFile = mock(VirtualFile.class);
        when(mockEvent.getProject()).thenReturn(null);
        when(mockEvent.getData(CommonDataKeys.VIRTUAL_FILE)).thenReturn(crdFile);
        
        try (MockedConstruction<CtyService> mockedService = mockConstruction(CtyService.class)) {
            action.actionPerformed(mockEvent);
            
            assertEquals("Should not create CtyService instance", 0, mockedService.constructed().size());
        }
    }
    
    @Test
    public void testActionPerformed_NoFile_DoesNotCallService() {
        when(mockEvent.getProject()).thenReturn(getProject());
        when(mockEvent.getData(CommonDataKeys.VIRTUAL_FILE)).thenReturn(null);
        
        try (MockedConstruction<CtyService> mockedService = mockConstruction(CtyService.class)) {
            action.actionPerformed(mockEvent);
            
            assertEquals("Should not create CtyService instance", 0, mockedService.constructed().size());
        }
    }
}