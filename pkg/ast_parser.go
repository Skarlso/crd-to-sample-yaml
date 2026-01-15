package pkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	conditionRegex = regexp.MustCompile(`\+cty:condition:for:(\w+)`)
	reasonRegex    = regexp.MustCompile(`\+cty:reason:for:(\w+)/(\w+)`)
)

// ConditionInfo represents a parsed condition annotation.
type ConditionInfo struct {
	CRDName     string
	Type        string
	Description string
	Reasons     []ReasonInfo
}

// ReasonInfo represents a parsed reason annotation.
type ReasonInfo struct {
	Name        string
	Description string
	Value       string
}

// ConditionParser handles parsing of condition annotations from Go source files.
type ConditionParser struct {
	fileSet    *token.FileSet
	conditions map[string]*ConditionInfo // key: CRDName/ConditionType
	reasons    map[string]*ReasonInfo    // key: CRDName/ConditionType/ReasonName
}

// NewConditionParser creates a new condition parser.
func NewConditionParser() *ConditionParser {
	return &ConditionParser{
		fileSet:    token.NewFileSet(),
		conditions: make(map[string]*ConditionInfo),
		reasons:    make(map[string]*ReasonInfo),
	}
}

// ParseGoFiles parses Go source files in the given directory for condition annotations.
// It recursively searches through all subdirectories.
func (p *ConditionParser) ParseGoFiles(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip non-Go files
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		if err := p.parseFile(path); err != nil {
			return fmt.Errorf("failed to parse file %s: %w", path, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	p.associateReasons()

	return nil
}

// parseFile parses a single Go source file.
func (p *ConditionParser) parseFile(filename string) error {
	src, err := parser.ParseFile(p.fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	ast.Inspect(src, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.GenDecl:
			p.parseGenDecl(n)
		case *ast.ValueSpec:
			p.parseValueSpec(n)
		}

		return true
	})

	return nil
}

// parseGenDecl parses general declarations (const, var, type).
func (p *ConditionParser) parseGenDecl(genDecl *ast.GenDecl) {
	if genDecl.Doc == nil {
		return
	}

	for _, spec := range genDecl.Specs {
		switch s := spec.(type) { //nolint:gocritic // type switch
		case *ast.ValueSpec:
			p.parseValueSpecWithComments(s, genDecl.Doc)
		}
	}
}

// parseValueSpec parses value specifications (const/var declarations).
func (p *ConditionParser) parseValueSpec(spec *ast.ValueSpec) {
	if spec.Doc == nil {
		return
	}

	p.parseValueSpecWithComments(spec, spec.Doc)
}

// parseValueSpecWithComments parses value spec with associated comments.
func (p *ConditionParser) parseValueSpecWithComments(spec *ast.ValueSpec, commentGroup *ast.CommentGroup) {
	comments := extractComments(commentGroup)

	for i, name := range spec.Names {
		var value string

		if i < len(spec.Values) {
			if lit, ok := spec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
				// Remove quotes from string literal if exists
				value = strings.Trim(lit.Value, `"`)
			}
		}

		p.parseAnnotations(name.Name, value, comments)
	}
}

// parseAnnotations parses condition and reason annotations from comments.
func (p *ConditionParser) parseAnnotations(varName, value string, comments []string) {
	var (
		description   strings.Builder
		crdName       string
		isCondition   bool
		reasonTargets []struct{ crdName, conditionType string } // Store multiple reason annotations
	)

	for _, comment := range comments {
		if matches := conditionRegex.FindStringSubmatch(comment); len(matches) > 1 {
			isCondition = true
			crdName = matches[1]

			continue
		}

		if matches := reasonRegex.FindStringSubmatch(comment); len(matches) > 2 { //nolint:mnd // crdName/conditionType
			// Collect all reason annotations instead of just keeping the last one
			reasonTargets = append(reasonTargets, struct{ crdName, conditionType string }{
				crdName:       matches[1],
				conditionType: matches[2],
			})

			continue
		}

		if !strings.Contains(comment, "+cty:") {
			if description.Len() > 0 {
				description.WriteString("\n")
			}

			description.WriteString(strings.TrimSpace(comment))
		}
	}

	// create "condition" if annotation was found
	if isCondition {
		key := fmt.Sprintf("%s/%s", crdName, varName)

		conditionValue := value
		if conditionValue == "" {
			conditionValue = varName // fallback to variable name if no value
		}

		p.conditions[key] = &ConditionInfo{
			CRDName:     crdName,
			Type:        conditionValue,
			Description: strings.TrimSpace(description.String()),
			Reasons:     []ReasonInfo{},
		}
	}

	// create "reason" entries for each annotation found
	for _, target := range reasonTargets {
		key := fmt.Sprintf("%s/%s/%s", target.crdName, target.conditionType, varName)

		reasonValue := value
		if reasonValue == "" {
			reasonValue = varName // fallback to variable name if no value
		}

		p.reasons[key] = &ReasonInfo{
			Name:        reasonValue,
			Description: strings.TrimSpace(description.String()),
			Value:       reasonValue,
		}
	}
}

// associateReasons associates parsed reasons with their corresponding conditions.
func (p *ConditionParser) associateReasons() {
	for reasonKey, reason := range p.reasons {
		parts := strings.Split(reasonKey, "/")
		if len(parts) != 3 { //nolint:mnd // crdName/conditionType/reasonName
			continue
		}

		crdName := parts[0]
		conditionRef := parts[1] // This could be variable name or condition value from the annotation

		for conditionKey, condition := range p.conditions {
			if condition.CRDName == crdName {
				// Match based on either:
				// 1. The condition key which includes the variable name
				// 2. The condition Type (value)
				conditionParts := strings.Split(conditionKey, "/")
				if len(conditionParts) == 2 && (conditionParts[1] == conditionRef || condition.Type == conditionRef) {
					condition.Reasons = append(condition.Reasons, *reason)
					// Don't break - continue to associate with all matching conditions
				}
			}
		}
	}
}

// GetConditions returns all parsed conditions grouped by CRD name.
// Conditions and their reasons are sorted for deterministic output.
func (p *ConditionParser) GetConditions() map[string][]ConditionInfo {
	result := make(map[string][]ConditionInfo)

	for _, condition := range p.conditions {
		sort.Slice(condition.Reasons, func(i, j int) bool {
			return condition.Reasons[i].Name < condition.Reasons[j].Name
		})
		result[condition.CRDName] = append(result[condition.CRDName], *condition)
	}

	for crdName := range result {
		sort.Slice(result[crdName], func(i, j int) bool {
			return result[crdName][i].Type < result[crdName][j].Type
		})
	}

	return result
}

// extractComments extracts comment text from a comment group.
func extractComments(cg *ast.CommentGroup) []string {
	if cg == nil {
		return nil
	}

	var comments []string

	for _, comment := range cg.List {
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")

		text = strings.TrimSpace(text)
		if text != "" {
			comments = append(comments, text)
		}
	}

	return comments
}
