package com.skarlso.crdtosampleyaml.services;

import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.psi.PsiFile;
import com.intellij.psi.PsiManager;
import org.yaml.snakeyaml.Yaml;
import org.yaml.snakeyaml.constructor.SafeConstructor;
import org.yaml.snakeyaml.LoaderOptions;

import java.io.StringReader;
import java.util.Map;

public class CrdDetectorService {
    
    private final Project project;
    private final Yaml yaml;
    
    public CrdDetectorService(Project project) {
        this.project = project;
        this.yaml = new Yaml(new SafeConstructor(new LoaderOptions()));
    }
    
    public boolean isCrdFile(VirtualFile file) {
        if (file == null || !isYamlFile(file)) {
            return false;
        }
        
        try {
            PsiFile psiFile = PsiManager.getInstance(project).findFile(file);
            if (psiFile == null) {
                return false;
            }
            
            String content = psiFile.getText();
            return isCrdContent(content);
        } catch (Exception e) {
            return false;
        }
    }
    
    public boolean isCrdContent(String content) {
        if (content == null || content.trim().isEmpty()) {
            return false;
        }
        
        try {
            Object document = yaml.load(new StringReader(content));
            if (!(document instanceof Map)) {
                return false;
            }
            
            @SuppressWarnings("unchecked")
            Map<String, Object> yamlMap = (Map<String, Object>) document;
            
            // Check for CRD identifying fields
            String kind = (String) yamlMap.get("kind");
            String apiVersion = (String) yamlMap.get("apiVersion");
            
            boolean isCrd = "CustomResourceDefinition".equals(kind) && 
                           apiVersion != null && 
                           apiVersion.startsWith("apiextensions.k8s.io/");
            
            // Debug logging to help troubleshoot
            if (!isCrd) {
                System.out.println("CRD Detection Debug - Kind: " + kind + ", ApiVersion: " + apiVersion);
            }
            
            return isCrd;
        } catch (Exception e) {
            System.out.println("CRD Detection Error: " + e.getMessage());
            return false;
        }
    }
    
    public String extractCrdName(VirtualFile file) {
        if (!isCrdFile(file)) {
            return null;
        }
        
        try {
            PsiFile psiFile = PsiManager.getInstance(project).findFile(file);
            if (psiFile == null) {
                return null;
            }
            
            String content = psiFile.getText();
            Object document = yaml.load(new StringReader(content));
            
            if (!(document instanceof Map)) {
                return null;
            }
            
            @SuppressWarnings("unchecked")
            Map<String, Object> yamlMap = (Map<String, Object>) document;
            
            @SuppressWarnings("unchecked")
            Map<String, Object> metadata = (Map<String, Object>) yamlMap.get("metadata");
            if (metadata != null) {
                return (String) metadata.get("name");
            }
            
            return null;
        } catch (Exception e) {
            return null;
        }
    }
    
    private boolean isYamlFile(VirtualFile file) {
        String extension = file.getExtension();
        return "yaml".equalsIgnoreCase(extension) || "yml".equalsIgnoreCase(extension);
    }
}