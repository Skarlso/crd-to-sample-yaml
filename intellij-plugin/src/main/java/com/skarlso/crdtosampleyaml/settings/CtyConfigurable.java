package com.skarlso.crdtosampleyaml.settings;

import com.intellij.openapi.fileChooser.FileChooserDescriptor;
import com.intellij.openapi.options.Configurable;
import com.intellij.openapi.ui.TextFieldWithBrowseButton;
import com.intellij.ui.components.JBCheckBox;
import com.intellij.ui.components.JBTextField;
import com.intellij.util.ui.FormBuilder;
import org.jetbrains.annotations.Nls;
import org.jetbrains.annotations.Nullable;

import javax.swing.*;

public class CtyConfigurable implements Configurable {
    
    private JBTextField ctyPathField;
    private TextFieldWithBrowseButton customOutputPathField;
    private JBCheckBox showNotificationsCheckBox;
    private JComboBox<String> outputLocationComboBox;
    private JPanel mainPanel;
    
    @Nls(capitalization = Nls.Capitalization.Title)
    @Override
    public String getDisplayName() {
        return "CRD to Sample YAML";
    }
    
    @Nullable
    @Override
    public JComponent createComponent() {
        ctyPathField = new JBTextField();
        ctyPathField.getEmptyText().setText("Path to cty binary (leave empty to use PATH)");
        
        customOutputPathField = new TextFieldWithBrowseButton();
        customOutputPathField.addBrowseFolderListener(
            "Select Output Directory",
            "Choose directory for generated sample files",
            null,
            new FileChooserDescriptor(false, true, false, false, false, false)
        );
        
        showNotificationsCheckBox = new JBCheckBox("Show notifications");
        
        outputLocationComboBox = new JComboBox<>(new String[]{
            "same_directory",
            "custom_directory"
        });
        
        mainPanel = FormBuilder.createFormBuilder()
            .addLabeledComponent("CTY Binary Path:", ctyPathField, 1, false)
            .addSeparator()
            .addLabeledComponent("Output Location:", outputLocationComboBox, 1, false)
            .addLabeledComponent("Custom Output Directory:", customOutputPathField, 1, false)
            .addSeparator()
            .addComponent(showNotificationsCheckBox, 1)
            .addComponentFillVertically(new JPanel(), 0)
            .getPanel();
        
        // Enable/disable custom path field based on combo box selection
        outputLocationComboBox.addActionListener(e -> {
            boolean isCustom = "custom_directory".equals(outputLocationComboBox.getSelectedItem());
            customOutputPathField.setEnabled(isCustom);
        });
        
        return mainPanel;
    }
    
    @Override
    public boolean isModified() {
        CtySettings settings = CtySettings.getInstance();
        
        return !ctyPathField.getText().equals(settings.getCtyPath()) ||
               showNotificationsCheckBox.isSelected() != settings.isShowNotifications() ||
               !outputLocationComboBox.getSelectedItem().equals(settings.getOutputLocation()) ||
               !customOutputPathField.getText().equals(settings.getCustomOutputPath());
    }
    
    @Override
    public void apply() {
        CtySettings settings = CtySettings.getInstance();
        
        settings.setCtyPath(ctyPathField.getText());
        settings.setShowNotifications(showNotificationsCheckBox.isSelected());
        settings.setOutputLocation((String) outputLocationComboBox.getSelectedItem());
        settings.setCustomOutputPath(customOutputPathField.getText());
    }
    
    @Override
    public void reset() {
        CtySettings settings = CtySettings.getInstance();
        
        ctyPathField.setText(settings.getCtyPath());
        showNotificationsCheckBox.setSelected(settings.isShowNotifications());
        outputLocationComboBox.setSelectedItem(settings.getOutputLocation());
        customOutputPathField.setText(settings.getCustomOutputPath());
        
        // Update custom path field state
        boolean isCustom = "custom_directory".equals(settings.getOutputLocation());
        customOutputPathField.setEnabled(isCustom);
    }
}