package com.skarlso.crdtosampleyaml.services;

import com.intellij.execution.configurations.GeneralCommandLine;
import com.intellij.execution.process.ProcessOutput;
import com.intellij.openapi.vfs.VirtualFile;
import com.intellij.testFramework.fixtures.BasePlatformTestCase;
import com.skarlso.crdtosampleyaml.settings.CtySettings;
import org.junit.Test;
import org.mockito.MockedStatic;

import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.mockStatic;
import static org.mockito.Mockito.when;

/**
 * These cover the command that actually gets handed to cty. Running the service end to end
 * would mean driving a background task and a progress dialog, which the headless test fixture
 * has no UI for; the invocation itself is where the interesting behaviour lives.
 */
public class CtyServiceTest extends BasePlatformTestCase {

    private static final String CTY = "/usr/local/bin/cty";

    private CtyService ctyService;
    private CtySettings settings;

    @Override
    protected void setUp() throws Exception {
        super.setUp();
        ctyService = new CtyService(getProject());
        settings = new CtySettings();
    }

    private VirtualFile crdFile() {
        return myFixture.createFile("test-crd.yaml", """
            apiVersion: apiextensions.k8s.io/v1
            kind: CustomResourceDefinition
            metadata:
              name: myresources.example.com
            """);
    }

    @Test
    public void testGenerateCommand_Complete() {
        VirtualFile crd = crdFile();

        GeneralCommandLine command = ctyService.buildGenerateCommand(
                CTY, crd, CtyService.GenerationType.COMPLETE, "/tmp/out");

        assertEquals(CTY, command.getExePath());
        assertEquals(
                java.util.List.of("generate", "crd", "-c", crd.getPath(), "-o", "/tmp/out"),
                command.getParametersList().getList());
    }

    @Test
    public void testGenerateCommand_MinimalUsesMinimalFlag() {
        GeneralCommandLine command = ctyService.buildGenerateCommand(
                CTY, crdFile(), CtyService.GenerationType.MINIMAL, "/tmp/out");

        assertTrue("minimal generation must pass -l",
                command.getParametersList().getList().contains("-l"));
    }

    @Test
    public void testGenerateCommand_CommentsUsesCommentsFlag() {
        GeneralCommandLine command = ctyService.buildGenerateCommand(
                CTY, crdFile(), CtyService.GenerationType.WITH_COMMENTS, "/tmp/out");

        assertTrue("commented generation must pass -m",
                command.getParametersList().getList().contains("-m"));
    }

    /**
     * This used to build "validate -c crd -s sample", which cty rejects outright because
     * validate had no -s flag. It has to target the sample subcommand.
     */
    @Test
    public void testValidateCommand_TargetsSampleSubcommand() {
        VirtualFile crd = crdFile();
        VirtualFile sample = myFixture.createFile("sample.yaml", """
            apiVersion: example.com/v1
            kind: MyResource
            metadata:
              name: test-resource
            """);

        GeneralCommandLine command = ctyService.buildValidateCommand(CTY, sample, crd);

        assertEquals(CTY, command.getExePath());
        assertEquals(
                java.util.List.of("validate", "sample", "-c", crd.getPath(), "-s", sample.getPath()),
                command.getParametersList().getList());
    }

    @Test
    public void testOutputDirectory_DefaultsToCrdDirectory() {
        VirtualFile crd = crdFile();

        try (MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            settingsMock.when(CtySettings::getInstance).thenReturn(settings);

            assertEquals(crd.getParent().getPath(), ctyService.resolveOutputDirectory(crd));
        }
    }

    @Test
    public void testOutputDirectory_HonoursConfiguredCustomDirectory() {
        VirtualFile crd = crdFile();
        settings.setOutputLocation(CtySettings.OUTPUT_CUSTOM_DIRECTORY);
        settings.setCustomOutputPath("/tmp/samples");

        try (MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            settingsMock.when(CtySettings::getInstance).thenReturn(settings);

            assertEquals("/tmp/samples", ctyService.resolveOutputDirectory(crd));
        }
    }

    @Test
    public void testOutputDirectory_FallsBackWhenCustomDirectoryIsBlank() {
        VirtualFile crd = crdFile();
        settings.setOutputLocation(CtySettings.OUTPUT_CUSTOM_DIRECTORY);
        settings.setCustomOutputPath("   ");

        try (MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            settingsMock.when(CtySettings::getInstance).thenReturn(settings);

            assertEquals(crd.getParent().getPath(), ctyService.resolveOutputDirectory(crd));
        }
    }

    /**
     * A bare "cty" was the old default for this setting and is not a usable file path,
     * so it must not be treated as one.
     */
    @Test
    public void testCtyPath_LegacyBareNameFallsBackToPathLookup() {
        settings.setCtyPath("cty");

        try (MockedStatic<CtySettings> settingsMock = mockStatic(CtySettings.class)) {
            settingsMock.when(CtySettings::getInstance).thenReturn(settings);

            // Either resolved from PATH or not found, but never the unusable literal.
            String resolved = ctyService.getCtyPath();
            assertFalse("bare name must not be returned as a path", "cty".equals(resolved));
        }
    }

    @Test
    public void testDescribeFailure_PrefersStderr() {
        ProcessOutput output = mock(ProcessOutput.class);
        when(output.getStderr()).thenReturn("Error: Invalid CRD format");

        assertEquals("Error: Invalid CRD format", ctyService.describeFailure(output));
    }

    @Test
    public void testDescribeFailure_FallsBackToStdout() {
        ProcessOutput output = mock(ProcessOutput.class);
        when(output.getStderr()).thenReturn("");
        when(output.getStdout()).thenReturn("sample is not valid");

        assertEquals("sample is not valid", ctyService.describeFailure(output));
    }

    @Test
    public void testDescribeFailure_FallsBackToExitCode() {
        ProcessOutput output = mock(ProcessOutput.class);
        when(output.getStderr()).thenReturn("");
        when(output.getStdout()).thenReturn("");
        when(output.getExitCode()).thenReturn(2);

        assertTrue(ctyService.describeFailure(output).contains("2"));
    }

    @Test
    public void testDescribeFailure_HandlesProcessThatNeverStarted() {
        assertNotNull(ctyService.describeFailure(null));
    }
}
