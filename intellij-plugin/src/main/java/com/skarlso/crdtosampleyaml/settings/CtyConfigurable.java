package com.skarlso.crdtosampleyaml.settings;

import com.intellij.openapi.fileChooser.FileChooserDescriptorFactory;
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
            null,
            FileChooserDescriptorFactory.createSingleFolderDescriptor()
                .withTitle("Select Output Directory")
                .withDescription("Choose directory for generated sample files")
        );
        
        showNotificationsCheckBox = new JBCheckBox("Show notifications");
        
        outputLocationComboBox = new JComboBox<>(new String[]{
            CtySettings.OUTPUT_SAME_DIRECTORY,
            CtySettings.OUTPUT_CUSTOM_DIRECTORY
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
            boolean isCustom = CtySettings.OUTPUT_CUSTOM_DIRECTORY.equals(outputLocationComboBox.getSelectedItem());
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
        boolean isCustom = CtySettings.OUTPUT_CUSTOM_DIRECTORY.equals(settings.getOutputLocation());
        customOutputPathField.setEnabled(isCustom);
    }
}