package pkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
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
func (p *ConditionParser) ParseGoFiles(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}

	pattern := filepath.Join(dir, "*.go")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob Go files: %w", err)
	}

	for _, file := range files {
		if err := p.parseFile(file); err != nil {
			return fmt.Errorf("failed to parse file %s: %w", file, err)
		}
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
		description            strings.Builder
		crdName, conditionType string
		isCondition, isReason  bool
	)

	for _, comment := range comments {
		if matches := conditionRegex.FindStringSubmatch(comment); len(matches) > 1 {
			isCondition = true
			crdName = matches[1]

			continue
		}

		if matches := reasonRegex.FindStringSubmatch(comment); len(matches) > 2 { //nolint:mnd // crdName/conditionType
			isReason = true
			crdName = matches[1]
			conditionType = matches[2]

			continue
		}

		if !strings.Contains(comment, "+cty:") {
			if description.Len() > 0 {
				description.WriteString(" ")
			}

			description.WriteString(strings.TrimSpace(comment))
		}
	}

	// create "condition" if annotation was found
	if isCondition {
		key := fmt.Sprintf("%s/%s", crdName, varName)
		p.conditions[key] = &ConditionInfo{
			CRDName:     crdName,
			Type:        varName,
			Description: strings.TrimSpace(description.String()),
			Reasons:     []ReasonInfo{},
		}
	}

	// create "reason" if annotation was found
	if isReason {
		key := fmt.Sprintf("%s/%s/%s", crdName, conditionType, varName)
		p.reasons[key] = &ReasonInfo{
			Name:        varName,
			Description: strings.TrimSpace(description.String()),
			Value:       value,
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
		conditionType := parts[1]

		for _, condition := range p.conditions {
			if condition.CRDName == crdName {
				if strings.Contains(condition.Type, conditionType) {
					condition.Reasons = append(condition.Reasons, *reason)

					break
				}
			}
		}
	}
}

// GetConditions returns all parsed conditions grouped by CRD name.
func (p *ConditionParser) GetConditions() map[string][]ConditionInfo {
	result := make(map[string][]ConditionInfo)

	for _, condition := range p.conditions {
		result[condition.CRDName] = append(result[condition.CRDName], *condition)
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
