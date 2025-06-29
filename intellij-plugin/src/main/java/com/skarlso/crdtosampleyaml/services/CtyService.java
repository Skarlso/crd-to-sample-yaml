package com.skarlso.crdtosampleyaml.services;

import com.intellij.execution.ExecutionException;
import com.intellij.execution.configurations.GeneralCommandLine;
import com.intellij.execution.process.ProcessHandler;
import com.intellij.execution.process.ProcessHandlerFactory;
import com.intellij.execution.process.ProcessOutput;
import com.intellij.execution.util.ExecUtil;
import com.intellij.notification.Notification;
import com.intellij.notification.NotificationDisplayType;
import com.intellij.notification.NotificationGroup;
import com.intellij.notification.NotificationType;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.skarlso.crdtosampleyaml.settings.CtySettings;

import java.io.File;
import java.nio.file.Path;
import java.nio.file.Paths;

public class CtyService {
    
    private static final NotificationGroup NOTIFICATION_GROUP = 
        NotificationGroup.balloonGroup("CRD to Sample YAML");
    
    private final Project project;
    
    public CtyService(Project project) {
        this.project = project;
    }
    
    public void generateSample(VirtualFile crdFile, GenerationType type) {
        try {
            String ctyPath = getCtyPath();
            if (ctyPath == null) {
                showError("CTY binary not found. Please install it or configure the path in settings.");
                return;
            }
            
            GeneralCommandLine commandLine = buildCommand(ctyPath, crdFile, type);
            ProcessOutput output = ExecUtil.execAndGetOutput(commandLine);
            
            if (output.getExitCode() == 0) {
                String outputPath = getOutputPath(crdFile, type);
                showSuccess("Sample YAML generated: " + outputPath);
            } else {
                showError("Failed to generate sample: " + output.getStderr());
            }
        } catch (ExecutionException e) {
            showError("Error executing CTY: " + e.getMessage());
        }
    }
    
    public void validateSample(VirtualFile sampleFile, VirtualFile crdFile) {
        try {
            String ctyPath = getCtyPath();
            if (ctyPath == null) {
                showError("CTY binary not found. Please install it or configure the path in settings.");
                return;
            }
            
            GeneralCommandLine commandLine = buildValidateCommand(ctyPath, sampleFile, crdFile);
            ProcessOutput output = ExecUtil.execAndGetOutput(commandLine);
            
            if (output.getExitCode() == 0) {
                showSuccess("Sample YAML is valid!");
            } else {
                showError("Validation failed: " + output.getStderr());
            }
        } catch (ExecutionException e) {
            showError("Error executing CTY validation: " + e.getMessage());
        }
    }
    
    private GeneralCommandLine buildCommand(String ctyPath, VirtualFile crdFile, GenerationType type) {
        GeneralCommandLine commandLine = new GeneralCommandLine();
        commandLine.setExePath(ctyPath);
        commandLine.addParameter("generate");
        commandLine.addParameter("crd");
        
        // Add CRD file parameter
        commandLine.addParameter("-c");
        commandLine.addParameter(crdFile.getPath());
        
        // Add type-specific flags
        switch (type) {
            case MINIMAL:
                commandLine.addParameter("-l");
                break;
            case WITH_COMMENTS:
                commandLine.addParameter("-m");
                break;
            case COMPLETE:
            default:
                // No additional flags for complete
                break;
        }
        
        // Set output directory to same as CRD file
        String outputDir = crdFile.getParent().getPath();
        commandLine.addParameter("-o");
        commandLine.addParameter(outputDir);
        
        return commandLine;
    }
    
    private GeneralCommandLine buildValidateCommand(String ctyPath, VirtualFile sampleFile, VirtualFile crdFile) {
        GeneralCommandLine commandLine = new GeneralCommandLine();
        commandLine.setExePath(ctyPath);
        commandLine.addParameter("validate");
        commandLine.addParameter("-c");
        commandLine.addParameter(crdFile.getPath());
        commandLine.addParameter("-s");
        commandLine.addParameter(sampleFile.getPath());
        
        return commandLine;
    }
    
    private String getCtyPath() {
        CtySettings settings = CtySettings.getInstance();
        String configuredPath = settings.getCtyPath();
        
        if (configuredPath != null && !configuredPath.trim().isEmpty()) {
            File file = new File(configuredPath);
            if (file.exists() && file.canExecute()) {
                return configuredPath;
            }
        }
        
        // Try to find in PATH
        try {
            ProcessOutput output = ExecUtil.execAndGetOutput(new GeneralCommandLine("which", "cty"));
            if (output.getExitCode() == 0) {
                return output.getStdout().trim();
            }
        } catch (ExecutionException e) {
            // Ignore and continue
        }
        
        return null;
    }
    
    private String getOutputPath(VirtualFile crdFile, GenerationType type) {
        // CTY generates files based on the Kind name from CRD, not filename
        // We'll show the directory and let the user discover the actual filename
        String outputDir = crdFile.getParent().getPath();
        return outputDir + "/*_sample.yaml (check output directory for generated file)";
    }
    
    private void showSuccess(String message) {
        Notification notification = NOTIFICATION_GROUP.createNotification(
            "CRD to Sample YAML", 
            message, 
            NotificationType.INFORMATION
        );
        notification.notify(project);
    }
    
    private void showError(String message) {
        Notification notification = NOTIFICATION_GROUP.createNotification(
            "CRD to Sample YAML", 
            message, 
            NotificationType.ERROR
        );
        notification.notify(project);
    }
    
    public enum GenerationType {
        COMPLETE,
        MINIMAL,
        WITH_COMMENTS
    }
}