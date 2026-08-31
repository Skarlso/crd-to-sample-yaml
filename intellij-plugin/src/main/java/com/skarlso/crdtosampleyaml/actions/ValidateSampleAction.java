package com.skarlso.crdtosampleyaml.actions;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.fileChooser.FileChooser;
import com.intellij.openapi.fileChooser.FileChooserDescriptor;
import com.intellij.openapi.fileChooser.FileChooserDescriptorFactory;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.skarlso.crdtosampleyaml.services.CtyService;

public class ValidateSampleAction extends BaseAction {

    @Override
    public void actionPerformed(AnActionEvent e) {
        Project project = getProject(e);
        VirtualFile crdFile = getSelectedFile(e);

        if (project == null || crdFile == null) {
            return;
        }

        FileChooserDescriptor descriptor = FileChooserDescriptorFactory.singleFile()
                .withTitle("Select Sample YAML to Validate")
                .withDescription("Choose the sample YAML file to validate against the CRD")
                .withFileFilter(file -> {
                    String extension = file.getExtension();

                    return "yaml".equalsIgnoreCase(extension) || "yml".equalsIgnoreCase(extension);
                });

        VirtualFile sampleFile = FileChooser.chooseFile(descriptor, project, crdFile.getParent());
        if (sampleFile != null) {
            new CtyService(project).validateSample(sampleFile, crdFile);
        }
    }
}
