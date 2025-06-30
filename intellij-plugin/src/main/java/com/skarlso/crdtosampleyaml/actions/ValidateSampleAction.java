package com.skarlso.crdtosampleyaml.actions;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.fileChooser.FileChooser;
import com.intellij.openapi.fileChooser.FileChooserDescriptor;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.skarlso.crdtosampleyaml.services.CtyService;

public class ValidateSampleAction extends BaseAction {
    
    @Override
    public void actionPerformed(AnActionEvent e) {
        Project project = getProject(e);
        VirtualFile crdFile = getSelectedFile(e);
        
        if (project != null && crdFile != null) {
            // Show file chooser to select the sample YAML to validate
            FileChooserDescriptor descriptor = new FileChooserDescriptor(
                true, false, false, false, false, false
            );
            descriptor.setTitle("Select Sample YAML to Validate");
            descriptor.setDescription("Choose the sample YAML file to validate against the CRD");
            descriptor.withFileFilter(file -> {
                String extension = file.getExtension();
                return "yaml".equalsIgnoreCase(extension) || "yml".equalsIgnoreCase(extension);
            });
            
            VirtualFile sampleFile = FileChooser.chooseFile(descriptor, project, crdFile.getParent());
            if (sampleFile != null) {
                CtyService ctyService = new CtyService(project);
                ctyService.validateSample(sampleFile, crdFile);
            }
        }
    }
}