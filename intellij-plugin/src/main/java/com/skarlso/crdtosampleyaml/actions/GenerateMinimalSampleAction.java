package com.skarlso.crdtosampleyaml.actions;

import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.skarlso.crdtosampleyaml.services.CtyService;

public class GenerateMinimalSampleAction extends BaseAction {
    
    @Override
    public void actionPerformed(AnActionEvent e) {
        Project project = getProject(e);
        VirtualFile file = getSelectedFile(e);
        
        if (project != null && file != null) {
            CtyService ctyService = new CtyService(project);
            ctyService.generateSample(file, CtyService.GenerationType.MINIMAL);
        }
    }
}