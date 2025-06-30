package com.skarlso.crdtosampleyaml.services;

import com.intellij.execution.process.ProcessOutput;
import org.junit.Test;

import static org.mockito.Mockito.*;

/**
 * Simple compilation test to verify our mocking approach works
 */
public class CompilationTest {
    
    @Test
    public void testProcessOutputMocking() {
        ProcessOutput mockOutput = mock(ProcessOutput.class);
        when(mockOutput.getExitCode()).thenReturn(0);
        when(mockOutput.getStdout()).thenReturn("test output");
        when(mockOutput.getStderr()).thenReturn("test error");
        
        // This test just verifies our mocking approach compiles correctly
        assert mockOutput.getExitCode() == 0;
        assert "test output".equals(mockOutput.getStdout());
        assert "test error".equals(mockOutput.getStderr());
    }
}