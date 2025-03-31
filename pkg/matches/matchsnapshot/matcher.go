package matchsnapshot

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/matches"
	"github.com/Skarlso/crd-to-sample-yaml/pkg/tests"
)

// MatcherName is the name of this matcher YAML.
const MatcherName = "matchSnapshot"

// Config contains configuration details for the Matcher. Path, ignoring errors and minimal setting.
type Config struct {
	Path         string   `yaml:"path"`
	IgnoreErrors []string `yaml:"ignoreErrors,omitempty"`
	Minimal      bool     `yaml:"minimal"`
}

// Matcher is a snapshot based matcher.
type Matcher struct {
	Updater Updater
}

func init() {
	tests.Register(&Matcher{
		Updater: &Update{},
	}, MatcherName)
}

// Match actually does the matching.
func (m *Matcher) Match(ctx context.Context, crdLocation string, payload []byte) error {
	c := Config{}
	if err := yaml.Unmarshal(payload, &c); err != nil {
		return err
	}

	// we only create the snapshots if update is requested, otherwise,
	// we just loop check existing snapshots
	if v := ctx.Value(matches.UpdateSnapshotKey); v != nil {
		if err := m.Updater.Update(crdLocation, c.Path, c.Minimal); err != nil {
			return fmt.Errorf("failed to update snapshot at %s: %w", c.Path, err)
		}
	}

	var snapshots []string
	err := filepath.Walk(c.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip reading folders
		if info.IsDir() {
			return nil
		}

		// make sure we only check the snapshots that belong to this crd being checked.
		baseCrdName := strings.Trim(filepath.Base(crdLocation), filepath.Ext(crdLocation))
		if strings.Contains(filepath.Base(path), baseCrdName) {
			if filepath.Ext(path) == ".yaml" {
				if c.Minimal {
					// only check files that have the min extension.
					if strings.HasSuffix(filepath.Base(path), ".min.yaml") {
						snapshots = append(snapshots, path)
					}
				} else if !strings.HasSuffix(filepath.Base(path), ".min.yaml") {
					// only add the file if it specifically does NOT contain the min extension.
					snapshots = append(snapshots, path)
				}
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	content, err := os.ReadFile(filepath.Clean(crdLocation))
	if err != nil {
		return fmt.Errorf("failed to read source template: %w", err)
	}

	// gather all the errors for all the files
	var validationErrors error
	for _, s := range snapshots {
		// one snapshot will contain a single version and the validation
		// will know which version to check against
		snapshotContent, err := os.ReadFile(filepath.Clean(s))
		if err != nil {
			return fmt.Errorf("failed to read snapshot template: %w", err)
		}

		if err := matches.Validate(content, snapshotContent, c.IgnoreErrors); err != nil {
			validationErrors = errors.Join(validationErrors, err)
		}
	}

	return validationErrors
}
