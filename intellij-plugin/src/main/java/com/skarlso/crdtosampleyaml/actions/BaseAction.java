package com.skarlso.crdtosampleyaml.actions;

import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.CommonDataKeys;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.skarlso.crdtosampleyaml.services.CrdDetectorService;

public abstract class BaseAction extends AnAction {
    
    @Override
    public void update(AnActionEvent e) {
        Project project = e.getProject();
        VirtualFile file = e.getData(CommonDataKeys.VIRTUAL_FILE);
        
        boolean enabled = project != null && file != null && 
                         new CrdDetectorService(project).isCrdFile(file);
        
        e.getPresentation().setEnabledAndVisible(enabled);
    }
    
    protected VirtualFile getSelectedFile(AnActionEvent e) {
        return e.getData(CommonDataKeys.VIRTUAL_FILE);
    }
    
    protected Project getProject(AnActionEvent e) {
        return e.getProject();
    }
}