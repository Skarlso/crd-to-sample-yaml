package com.skarlso.crdtosampleyaml.settings;

import com.intellij.openapi.application.ApplicationManager;
import com.intellij.openapi.components.PersistentStateComponent;
import com.intellij.openapi.components.State;
import com.intellij.openapi.components.Storage;
import com.intellij.util.xmlb.XmlSerializerUtil;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;

@State(
    name = "CtySettings",
    storages = @Storage("ctySettings.xml")
)
public class CtySettings implements PersistentStateComponent<CtySettings> {
    
    public String ctyPath = "cty";
    public boolean showNotifications = true;
    public String outputLocation = "same_directory";
    public String customOutputPath = "";
    
    public static CtySettings getInstance() {
        return ApplicationManager.getApplication().getService(CtySettings.class);
    }
    
    @Nullable
    @Override
    public CtySettings getState() {
        return this;
    }
    
    @Override
    public void loadState(@NotNull CtySettings state) {
        XmlSerializerUtil.copyBean(state, this);
    }
    
    public String getCtyPath() {
        return ctyPath;
    }
    
    public void setCtyPath(String ctyPath) {
        this.ctyPath = ctyPath;
    }
    
    public boolean isShowNotifications() {
        return showNotifications;
    }
    
    public void setShowNotifications(boolean showNotifications) {
        this.showNotifications = showNotifications;
    }
    
    public String getOutputLocation() {
        return outputLocation;
    }
    
    public void setOutputLocation(String outputLocation) {
        this.outputLocation = outputLocation;
    }
    
    public String getCustomOutputPath() {
        return customOutputPath;
    }
    
    public void setCustomOutputPath(String customOutputPath) {
        this.customOutputPath = customOutputPath;
    }
}