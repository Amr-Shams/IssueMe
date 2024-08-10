package util

import (
	"bytes"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Define execCommand as a variable to be overridden in tests
var execCommand = exec.Command

type mockCmd struct {
	output []byte
	err    error
}

func (m *mockCmd) Output() ([]byte, error) {
	return m.output, m.err
}

func (m *mockCmd) Write(p []byte) (n int, err error) {
	return len(m.output), nil
}

func TestGetModifiedFiles(t *testing.T) {
	tests := []struct {
		name           string
		mockOutput     []byte
		mockError      error
		expectedResult []string
		expectedError  bool
	}{
		{
			name:           "No modified files",
			mockOutput:     []byte(""),
			mockError:      nil,
			expectedResult: []string{},
			expectedError:  false,
		},
		{
			name:           "Some modified files",
			mockOutput:     []byte(" M file1.txt\n M file2.txt"),
			mockError:      nil,
			expectedResult: []string{"file1.txt", "file2.txt"},
			expectedError:  false,
		},
		{
			name:           "Error case",
			mockOutput:     nil,
			mockError:      exec.ErrNotFound,
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the original execCommand and defer its restoration
			oldExecCommand := execCommand
			defer func() { execCommand = oldExecCommand }()

			// Replace execCommand with our mock
			execCommand = func(command string, args ...string) *exec.Cmd {
				return &exec.Cmd{
					Stdout: bytes.NewBuffer(tt.mockOutput),
					Stderr: &mockCmd{output: tt.mockOutput, err: tt.mockError},
				}
			}

			// Call the function
			result, err := GetModifiedFiles()

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}
