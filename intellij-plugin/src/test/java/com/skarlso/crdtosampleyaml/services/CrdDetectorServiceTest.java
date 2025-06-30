package com.skarlso.crdtosampleyaml.services;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.PsiFile;
import com.intellij.psi.PsiManager;
import com.intellij.testFramework.fixtures.BasePlatformTestCase;
import com.intellij.testFramework.fixtures.TempDirTestFixture;
import org.junit.Test;

public class CrdDetectorServiceTest extends BasePlatformTestCase {
    
    private CrdDetectorService crdDetectorService;
    
    @Override
    protected void setUp() throws Exception {
        super.setUp();
        crdDetectorService = new CrdDetectorService(getProject());
    }
    
    @Test
    public void testIsCrdContent_ValidCrd_ReturnsTrue() {
        String validCrdContent = """
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
            """;
        
        assertTrue("Should detect valid CRD content", 
                  crdDetectorService.isCrdContent(validCrdContent));
    }
    
    @Test
    public void testIsCrdContent_InvalidKind_ReturnsFalse() {
        String invalidContent = """
            apiVersion: v1
            kind: Pod
            metadata:
              name: test-pod
            spec:
              containers:
              - name: test
                image: nginx
            """;
        
        assertFalse("Should not detect Pod as CRD", 
                   crdDetectorService.isCrdContent(invalidContent));
    }
    
    @Test
    public void testIsCrdContent_InvalidApiVersion_ReturnsFalse() {
        String invalidContent = """
            apiVersion: v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """;
        
        assertFalse("Should not detect CRD with wrong API version", 
                   crdDetectorService.isCrdContent(invalidContent));
    }
    
    @Test
    public void testIsCrdContent_MultiDocument_ReturnsTrue() {
        String multiDocContent = """
            apiVersion: v1
            kind: Namespace
            metadata:
              name: test-namespace
            ---
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
            """;
        
        assertTrue("Should detect CRD in multi-document YAML", 
                  crdDetectorService.isCrdContent(multiDocContent));
    }
    
    @Test
    public void testIsCrdContent_EmptyContent_ReturnsFalse() {
        assertFalse("Should not detect empty content as CRD", 
                   crdDetectorService.isCrdContent(""));
        assertFalse("Should not detect null content as CRD", 
                   crdDetectorService.isCrdContent(null));
        assertFalse("Should not detect whitespace as CRD", 
                   crdDetectorService.isCrdContent("   \n\t  "));
    }
    
    @Test
    public void testIsCrdContent_InvalidYaml_ReturnsFalse() {
        String invalidYaml = """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            spec:
              - invalid: yaml: structure
                - more: invalid
            """;
        
        assertFalse("Should handle invalid YAML gracefully", 
                   crdDetectorService.isCrdContent(invalidYaml));
    }
    
    @Test
    public void testIsCrdFile_YamlFile_ChecksContent() throws Exception {
        VirtualFile yamlFile = myFixture.createFile("test-crd.yaml", """
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
        
        assertTrue("Should detect CRD file", crdDetectorService.isCrdFile(yamlFile));
    }
    
    @Test
    public void testIsCrdFile_NonYamlFile_ReturnsFalse() throws Exception {
        VirtualFile javaFile = myFixture.createFile("Test.java", "public class Test {}");
        
        assertFalse("Should not detect non-YAML file as CRD", 
                   crdDetectorService.isCrdFile(javaFile));
    }
    
    @Test
    public void testIsCrdFile_NullFile_ReturnsFalse() {
        assertFalse("Should handle null file gracefully", 
                   crdDetectorService.isCrdFile(null));
    }
    
    @Test
    public void testExtractCrdName_ValidCrd_ReturnsName() throws Exception {
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
        
        String name = crdDetectorService.extractCrdName(crdFile);
        assertEquals("Should extract CRD name from metadata", 
                    "myresources.example.com", name);
    }
    
    @Test
    public void testExtractCrdName_NonCrdFile_ReturnsNull() throws Exception {
        VirtualFile nonCrdFile = myFixture.createFile("pod.yaml", """
            apiVersion: v1
            kind: Pod
            metadata:
              name: test-pod
            spec:
              containers:
              - name: test
                image: nginx
            """);
        
        String name = crdDetectorService.extractCrdName(nonCrdFile);
        assertNull("Should return null for non-CRD file", name);
    }
}