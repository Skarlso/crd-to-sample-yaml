package com.skarlso.crdtosampleyaml.services;

import com.intellij.execution.ExecutionException;
import com.intellij.execution.configurations.GeneralCommandLine;
import com.intellij.execution.configurations.PathEnvironmentVariableUtil;
import com.intellij.execution.process.ProcessOutput;
import com.intellij.execution.util.ExecUtil;
import com.intellij.notification.NotificationGroupManager;
import com.intellij.notification.NotificationType;
import com.intellij.openapi.diagnostic.Logger;
import com.intellij.openapi.progress.ProgressIndicator;
import com.intellij.openapi.progress.Task;
import com.intellij.openapi.project.Project;
import com.intellij.openapi.vfs.LocalFileSystem;
import com.intellij.openapi.vfs.VirtualFile;
import com.skarlso.crdtosampleyaml.settings.CtySettings;

import java.io.File;

public class CtyService {

    private static final Logger LOG = Logger.getInstance(CtyService.class);
    private static final String NOTIFICATION_GROUP_ID = "CRD to Sample YAML";
    private static final String BINARY_NAME = "cty";

    private final Project project;

    public CtyService(Project project) {
        this.project = project;
    }

    public void generateSample(VirtualFile crdFile, GenerationType type) {
        runInBackground("Generating sample YAML", () -> {
            String ctyPath = getCtyPath();
            if (ctyPath == null) {
                showError(binaryMissingMessage());
                return;
            }

            String outputDir = resolveOutputDirectory(crdFile);
            ProcessOutput output = run(buildGenerateCommand(ctyPath, crdFile, type, outputDir));

            if (output != null && output.getExitCode() == 0) {
                refresh(outputDir);
                showInfo("Sample YAML generated in " + outputDir);
            } else {
                showError("Failed to generate sample: " + describeFailure(output));
            }
        });
    }

    public void validateSample(VirtualFile sampleFile, VirtualFile crdFile) {
        runInBackground("Validating sample YAML", () -> {
            String ctyPath = getCtyPath();
            if (ctyPath == null) {
                showError(binaryMissingMessage());
                return;
            }

            ProcessOutput output = run(buildValidateCommand(ctyPath, sampleFile, crdFile));

            if (output != null && output.getExitCode() == 0) {
                showInfo(sampleFile.getName() + " is valid.");
            } else {
                showError("Validation failed: " + describeFailure(output));
            }
        });
    }

    /**
     * Runs the work on a pooled thread. Generation shells out to cty, which must never happen
     * on the event dispatch thread or the whole IDE freezes for the duration of the process.
     */
    private void runInBackground(String title, Runnable work) {
        new Task.Backgroundable(project, title, true) {
            @Override
            public void run(ProgressIndicator indicator) {
                indicator.setIndeterminate(true);
                work.run();
            }
        }.queue();
    }

    private ProcessOutput run(GeneralCommandLine commandLine) {
        try {
            return ExecUtil.execAndGetOutput(commandLine);
        } catch (ExecutionException e) {
            LOG.warn("failed to execute " + commandLine.getCommandLineString(), e);
            return null;
        }
    }

    // Package-private so tests can assert on the exact cty invocation.
    GeneralCommandLine buildGenerateCommand(
            String ctyPath, VirtualFile crdFile, GenerationType type, String outputDir) {
        GeneralCommandLine commandLine = new GeneralCommandLine();
        commandLine.setExePath(ctyPath);
        commandLine.addParameters("generate", "crd");
        commandLine.addParameters("-c", crdFile.getPath());

        switch (type) {
            case MINIMAL:
                commandLine.addParameter("-l");
                break;
            case WITH_COMMENTS:
                commandLine.addParameter("-m");
                break;
            case COMPLETE:
            default:
                break;
        }

        commandLine.addParameters("-o", outputDir);

        return commandLine;
    }

    GeneralCommandLine buildValidateCommand(String ctyPath, VirtualFile sampleFile, VirtualFile crdFile) {
        GeneralCommandLine commandLine = new GeneralCommandLine();
        commandLine.setExePath(ctyPath);
        commandLine.addParameters("validate", "sample");
        commandLine.addParameters("-c", crdFile.getPath());
        commandLine.addParameters("-s", sampleFile.getPath());

        return commandLine;
    }

    /**
     * Where generated samples land: next to the CRD unless the settings point somewhere else.
     */
    String resolveOutputDirectory(VirtualFile crdFile) {
        CtySettings settings = CtySettings.getInstance();

        if (CtySettings.OUTPUT_CUSTOM_DIRECTORY.equals(settings.getOutputLocation())) {
            String custom = settings.getCustomOutputPath();
            if (custom != null && !custom.trim().isEmpty()) {
                return custom.trim();
            }
        }

        return crdFile.getParent().getPath();
    }

    String getCtyPath() {
        CtySettings settings = CtySettings.getInstance();
        String configuredPath = settings.getCtyPath();

        if (isExplicitPath(configuredPath)) {
            File file = new File(configuredPath.trim());
            if (file.isFile() && file.canExecute()) {
                return file.getAbsolutePath();
            }
        }

        // Resolves against PATH using the platform's own rules, so this also works on Windows
        // where `which` does not exist and the binary is cty.exe.
        File onPath = PathEnvironmentVariableUtil.findInPath(BINARY_NAME);

        return onPath != null ? onPath.getAbsolutePath() : null;
    }

    /**
     * A bare "cty" was the old default for this setting, so treat it as "look on PATH"
     * rather than as a relative file path that will never resolve.
     */
    private boolean isExplicitPath(String configuredPath) {
        return configuredPath != null
                && !configuredPath.trim().isEmpty()
                && !BINARY_NAME.equals(configuredPath.trim());
    }

    private String binaryMissingMessage() {
        String configured = CtySettings.getInstance().getCtyPath();

        if (isExplicitPath(configured)) {
            return "cty was not found at '" + configured.trim()
                    + "'. Fix the path in Settings | Tools | CRD to Sample YAML.";
        }

        return "cty was not found on your PATH. Install it (brew install crd-to-sample-yaml) "
                + "or set the full path in Settings | Tools | CRD to Sample YAML.";
    }

    /**
     * cty reports validation failures on stderr but falls back to stdout for some errors,
     * so prefer whichever actually carries the message.
     */
    String describeFailure(ProcessOutput output) {
        if (output == null) {
            return "cty could not be executed, see the IDE log for details.";
        }

        String stderr = output.getStderr().trim();
        if (!stderr.isEmpty()) {
            return stderr;
        }

        String stdout = output.getStdout().trim();
        if (!stdout.isEmpty()) {
            return stdout;
        }

        return "cty exited with code " + output.getExitCode() + ".";
    }

    /**
     * Generated files are written behind the VFS's back, so refresh or they stay invisible
     * in the project view until the next external change is noticed.
     */
    private void refresh(String outputDir) {
        VirtualFile dir = LocalFileSystem.getInstance().refreshAndFindFileByPath(outputDir);
        if (dir != null) {
            dir.refresh(true, false);
        }
    }

    private void showInfo(String message) {
        notify(message, NotificationType.INFORMATION);
    }

    private void showError(String message) {
        notify(message, NotificationType.ERROR);
    }

    private void notify(String message, NotificationType type) {
        // Errors are always worth surfacing; the setting only silences success balloons.
        if (type != NotificationType.ERROR && !CtySettings.getInstance().isShowNotifications()) {
            return;
        }

        NotificationGroupManager.getInstance()
                .getNotificationGroup(NOTIFICATION_GROUP_ID)
                .createNotification("CRD to Sample YAML", message, type)
                .notify(project);
    }

    public enum GenerationType {
        COMPLETE,
        MINIMAL,
        WITH_COMMENTS
    }
}
