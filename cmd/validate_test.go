package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCRD = `
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: samples.example.com
spec:
  group: example.com
  scope: Namespaced
  names:
    plural: samples
    singular: sample
    kind: Sample
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required:
                - image
              properties:
                image:
                  type: string
                replicas:
                  type: integer
`

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()

	loc := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(loc, []byte(content), 0o600))

	return loc
}

func TestRunSampleValidation(t *testing.T) {
	crdLoc := writeTempFile(t, "crd.yaml", testCRD)

	tests := []struct {
		name    string
		sample  string
		wantErr string
	}{
		{
			name: "valid sample passes",
			sample: `
apiVersion: example.com/v1
kind: Sample
metadata:
  name: valid
spec:
  image: nginx
  replicas: 3
`,
		},
		{
			name: "wrong field type is reported",
			sample: `
apiVersion: example.com/v1
kind: Sample
metadata:
  name: bad-type
spec:
  image: nginx
  replicas: "three"
`,
			wantErr: "spec.replicas in body must be of type integer",
		},
		{
			name: "missing required field is reported",
			sample: `
apiVersion: example.com/v1
kind: Sample
metadata:
  name: missing-required
spec:
  replicas: 1
`,
			wantErr: "spec.image in body is required",
		},
		{
			name: "unknown version is reported",
			sample: `
apiVersion: example.com/v99
kind: Sample
metadata:
  name: wrong-version
spec:
  image: nginx
`,
			wantErr: "not found amongst the available testing versions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args = &rootArgs{fileLocation: crdLoc}
			valArgs = &validateArgs{sample: writeTempFile(t, "sample.yaml", tt.sample)}

			err := runSampleValidation(sampleValidateCmd, nil)

			if tt.wantErr == "" {
				require.NoError(t, err)

				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestRunSampleValidationIgnoreErrors(t *testing.T) {
	crdLoc := writeTempFile(t, "crd.yaml", testCRD)
	sampleLoc := writeTempFile(t, "sample.yaml", `
apiVersion: example.com/v1
kind: Sample
metadata:
  name: missing-required
spec:
  replicas: 1
`)

	args = &rootArgs{fileLocation: crdLoc}
	valArgs = &validateArgs{sample: sampleLoc, ignoreErrors: []string{"spec.image in body is required"}}

	require.NoError(t, runSampleValidation(sampleValidateCmd, nil))
}

func TestCrdContentRequiresASource(t *testing.T) {
	_, err := crdContent(&rootArgs{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be set")
}

func TestDefaultOutputLocation(t *testing.T) {
	tests := []struct {
		name string
		args *rootArgs
		want string
	}{
		{
			name: "file input lands next to the CRD",
			args: &rootArgs{fileLocation: filepath.Join("some", "dir", "crd.yaml")},
			want: filepath.Join("some", "dir"),
		},
		{
			name: "folder input lands in that folder",
			args: &rootArgs{folderLocation: filepath.Join("some", "dir")},
			want: filepath.Join("some", "dir"),
		},
		{
			name: "remote input lands in the working directory",
			args: &rootArgs{url: "https://example.com/crd.yaml"},
			want: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, defaultOutputLocation(tt.args))
		})
	}
}
