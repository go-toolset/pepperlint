package core

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/go-toolset/pepperlint"
)

const deprecatedPrefix = `// Deprecated:`

// DeprecatedRule is a container for all deprecated rules
type DeprecatedRule struct {
	structRule *DeprecatedStructRule
	fieldRule  *DeprecatedFieldRule
	opRule     *DeprecatedOpRule
}

// NewDeprecatedRule will return a new set of deprecated rules
func NewDeprecatedRule(fset *token.FileSet) *DeprecatedRule {
	return &DeprecatedRule{
		structRule: NewDeprecatedStructRule(fset),
		fieldRule:  NewDeprecatedFieldRule(fset),
		opRule:     NewDeprecatedOpRule(fset),
	}
}

// AddRules will add rules for every deprecate rule
func (r *DeprecatedRule) AddRules(rules *pepperlint.Rules) {
	r.structRule.AddRules(rules)
	r.fieldRule.AddRules(rules)
	r.opRule.AddRules(rules)
}

// WithCache will add rules for every deprecate rule
func (r *DeprecatedRule) WithCache(cache *pepperlint.Cache) {
	r.structRule.WithCache(cache)
	r.fieldRule.WithCache(cache)
	r.opRule.WithCache(cache)
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
