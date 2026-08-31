package com.skarlso.crdtosampleyaml.actions;

import com.intellij.openapi.actionSystem.ActionUpdateThread;
import com.intellij.openapi.actionSystem.AnAction;
import com.intellij.openapi.actionSystem.AnActionEvent;
import com.intellij.openapi.actionSystem.CommonDataKeys;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.VirtualFile;
import com.skarlso.crdtosampleyaml.services.CrdDetectorService;
import org.jetbrains.annotations.NotNull;

public abstract class BaseAction extends AnAction {

    /**
     * update() parses the selected file to decide whether it is a CRD, which is far too much
     * work for the event dispatch thread. Since 2022.3 the platform requires this to be
     * declared explicitly, and it throws at runtime when an action overriding update() does not.
     */
    @Override
    public @NotNull ActionUpdateThread getActionUpdateThread() {
        return ActionUpdateThread.BGT;
    }

    @Override
    public void update(@NotNull AnActionEvent e) {
        VirtualFile file = e.getData(CommonDataKeys.VIRTUAL_FILE);

        boolean enabled = file != null && CrdDetectorService.getInstance().isCrdFile(file);

        e.getPresentation().setEnabledAndVisible(enabled);
    }

    protected VirtualFile getSelectedFile(AnActionEvent e) {
        return e.getData(CommonDataKeys.VIRTUAL_FILE);
    }

    protected Project getProject(AnActionEvent e) {
        return e.getProject();
    }
}
