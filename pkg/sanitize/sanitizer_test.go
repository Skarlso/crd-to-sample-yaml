package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitize(t *testing.T) {
	type args struct {
		content []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "sanitize",
			args: args{
				content: []byte(`test: {{ Value.something }}`),
			},
			want: []byte(`test: replaced`),
		},
		{
			name: "complex sanitize",
			args: args{
				content: []byte(`{{ if .something }}
test: {{ Value.something }}
normal: value
this:
  should:
    not:
      mess: {{ Value.should.work }}
      with: indentation
{{- end }}
`),
			},
			want: []byte(`test: replaced
normal: value
this:
  should:
    not:
      mess: replaced
      with: indentation
`),
		},
		{
			name: "begins with space",
			args: args{
				content: []byte(`{{ if .something }}
test: {{ Value.something }}
normal: value
this:
  should:
    not:
      mess: {{ Value.should.work }}
      with: indentation
  {{- end }}
`),
			},
			want: []byte(`test: replaced
normal: value
this:
  should:
    not:
      mess: replaced
      with: indentation
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sanitize(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sanitize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
