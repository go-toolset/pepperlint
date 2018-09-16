package core

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/go-toolset/pepperlint"
)

const deprecatedPrefix = `// Deprecated:`

// DeprecatedRule ...
type DeprecatedRule struct {
	structRule *DeprecatedStructRule
	fieldRule  *DeprecatedFieldRule
	opRule     *DeprecatedOpRule
}

// NewDeprecatedRule ...
func NewDeprecatedRule(fset *token.FileSet) *DeprecatedRule {
	return &DeprecatedRule{
		structRule: NewDeprecatedStructRule(fset),
		fieldRule:  NewDeprecatedFieldRule(fset),
		opRule:     NewDeprecatedOpRule(fset),
	}
}

// AddRules ...
func (r *DeprecatedRule) AddRules(v *pepperlint.Visitor) {
	r.structRule.AddRules(v)
	r.fieldRule.AddRules(v)
	r.opRule.AddRules(v)
}

// deprecatedFields will map what fields are deprecated in a struct type.
type deprecatedCache struct {
	KeyLookup   map[string]struct{}
	IndexLookup []bool
}

func (cache deprecatedCache) HasKey(key string) bool {
	_, ok := cache.KeyLookup[key]
	return ok
}

func getDeprecatedFields(fields *ast.FieldList) deprecatedCache {
	depFields := deprecatedCache{
		KeyLookup:   map[string]struct{}{},
		IndexLookup: make([]bool, len(fields.List)),
	}

	for i, field := range fields.List {
		if field.Doc == nil {
			continue
		}

		// only need to check the doc list because a golang comment
		// is defined to be a comment at the end of a statement.
		if len(field.Doc.List) == 0 {
			continue
		}

		// Check each line for '// Deprecated:'
		if hasDeprecatedComment(field.Doc) {
			for _, name := range field.Names {
				depFields.KeyLookup[name.Name] = struct{}{}
			}

			depFields.IndexLookup[i] = true
		}
	}

	return depFields
}

func hasDeprecatedComment(comments *ast.CommentGroup) bool {
	if comments == nil {
		return false
	}

	for _, comment := range comments.List {
		if strings.HasPrefix(comment.Text, deprecatedPrefix) {
			return true
		}
	}

	return false
}
