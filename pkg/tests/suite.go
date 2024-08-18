package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/yaml"

	"github.com/Skarlso/crd-to-sample-yaml/pkg/matches"
)

var matchers = map[string]matches.Matcher{}

// Register a matcher under a name.
func Register(matcher matches.Matcher, name string) {
	if _, ok := matchers[name]; ok {
		panic(fmt.Sprintf("a matcher with name %s already exists", name))
	}

	matchers[name] = matcher
}

// Runner can execute a suite of tests and return any none 0 exit statuses.
type Runner interface {
	Run() error
}

// NewSuiteRunner initializes the runner with a specific location to run tests from.
func NewSuiteRunner(location string, update bool) *SuiteRunner {
	return &SuiteRunner{
		Location: location,
		Update:   update,
		Updater:  &Update{},
	}
}

// SuiteRunner is a standard suite runner that runs suits sequentially.
type SuiteRunner struct {
	Location string
	Update   bool
	Updater  Updater
}

type Test struct {
	It      string                  `json:"it"`
	Asserts []*apiextensionsv1.JSON `json:"asserts"`
}

type Suite struct {
	Suite    string `json:"suite"`
	Tests    []Test `json:"tests"`
	Template string `json:"template"`
}

// Outcome is returned by a run.
type Outcome struct {
	Name     string
	Error    error
	Matcher  string
	Template string
}

// Run runs a suite of tests in sequence.
func (s *SuiteRunner) Run() ([]Outcome, error) {
	testMatrix, err := s.constructTestMatrix()
	if err != nil {
		return nil, fmt.Errorf("failed to construct test matrix: %w", err)
	}

	var outcome []Outcome

	for file, v := range testMatrix {
		for _, t := range v {
			for _, assert := range t.Asserts {
				m := map[string]*apiextensionsv1.JSON{}
				if err := yaml.Unmarshal(assert.Raw, &m); err != nil {
					outcome = append(outcome, Outcome{
						Name:     t.It,
						Error:    fmt.Errorf("yaml.Unmarshal() returned %w", err),
						Matcher:  "unknown",
						Template: file,
					})

					continue
				}
				for k, payload := range m {
					if _, ok := matchers[k]; !ok {
						outcome = append(outcome, Outcome{
							Name:     t.It,
							Error:    fmt.Errorf("matcher %s not registered", k),
							Matcher:  k,
							Template: file,
						})

						continue
					}

					if k == "matchSnapshot" && s.Update {
						if err := s.UpdateSnapshot(file, payload.Raw); err != nil {
							outcome = append(outcome, Outcome{
								Name:     t.It,
								Error:    fmt.Errorf("updater.Update() returned %w", err),
								Matcher:  k,
								Template: file,
							})

							continue
						}
					}

					matcher := matchers[k]
					if err := matcher.Match(file, payload.Raw); err != nil {
						// test failed
						outcome = append(outcome, Outcome{
							Name:     t.It,
							Error:    fmt.Errorf("matcher returned failure: %w", err),
							Matcher:  k,
							Template: file,
						})

						continue
					}

					// test passed
					outcome = append(outcome, Outcome{
						Name:     t.It,
						Matcher:  k,
						Template: file,
					})
				}
			}
		}
	}

	return outcome, nil
}

func (s *SuiteRunner) constructTestMatrix() (map[string][]Test, error) {
	var testFiles []string

	err := filepath.Walk(s.Location, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(filepath.Base(path), "_test.yaml") {
			testFiles = append(testFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("filepath.Walk() returned %w", err)
	}

	// build up the test matrix
	// for each template, gather the `tests`s
	testMatrix := map[string][]Test{}
	for _, testFile := range testFiles {
		content, err := os.ReadFile(testFile)
		if err != nil {
			return nil, fmt.Errorf("ioutil.ReadFile() returned %w", err)
		}

		suite := &Suite{}
		if err := yaml.Unmarshal(content, suite); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal() returned %w", err)
		}

		testMatrix[suite.Template] = suite.Tests
	}

	return testMatrix, nil
}

func (s *SuiteRunner) UpdateSnapshot(file string, payload []byte) error {
	var snapshotLocation struct {
		Name    string
		Minimal bool
	}
	if err := yaml.Unmarshal(payload, &snapshotLocation); err != nil {
		return fmt.Errorf("yaml.Unmarshal() returned %w", err)
	}

	if err := s.Updater.Update(file, snapshotLocation.Name, snapshotLocation.Minimal); err != nil {
		return fmt.Errorf("updater.Update() returned %w", err)
	}

	return nil
}
