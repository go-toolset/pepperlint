package pepperlint

// Rules contain a set of all rule types that will be ran
// during visitation.
type Rules struct {
	PackageRules  PackageRules
	FileRules     FileRules
	TypeSpecRules TypeSpecRules

	// Declarations
	GenDeclRules  GenDeclRules
	FuncDeclRules FuncDeclRules

	// Expressions
	CallExprRules CallExprRules

	// Statements
	AssignStmtRules AssignStmtRules
	BlockStmtRules  BlockStmtRules
	ReturnStmtRules ReturnStmtRules

	// Primitive Types

	// Complex Types
	StructTypeRules    StructTypeRules
	FieldRules         FieldRules
	FieldListRules     FieldListRules
	FuncTypeRules      FuncTypeRules
	InterfaceTypeRules InterfaceTypeRules

	// Container Types
	ArrayTypeRules ArrayTypeRules
	ChanTypeRules  ChanTypeRules
	MapTypeRules   MapTypeRules
}

// Merge will merge two rule sets together.
func (r *Rules) Merge(otherRules Rules) *Rules {
	r.PackageRules = append(r.PackageRules, otherRules.PackageRules...)
	r.FuncDeclRules = append(r.FuncDeclRules, otherRules.FuncDeclRules...)
	r.GenDeclRules = append(r.GenDeclRules, otherRules.GenDeclRules...)
	r.TypeSpecRules = append(r.TypeSpecRules, otherRules.TypeSpecRules...)
	r.AssignStmtRules = append(r.AssignStmtRules, otherRules.AssignStmtRules...)
	r.BlockStmtRules = append(r.BlockStmtRules, otherRules.BlockStmtRules...)
	r.StructTypeRules = append(r.StructTypeRules, otherRules.StructTypeRules...)
	r.FieldRules = append(r.FieldRules, otherRules.FieldRules...)
	r.FuncTypeRules = append(r.FuncTypeRules, otherRules.FuncTypeRules...)
	r.InterfaceTypeRules = append(r.InterfaceTypeRules, otherRules.InterfaceTypeRules...)
	r.ArrayTypeRules = append(r.ArrayTypeRules, otherRules.ArrayTypeRules...)
	r.ChanTypeRules = append(r.ChanTypeRules, otherRules.ChanTypeRules...)
	r.MapTypeRules = append(r.MapTypeRules, otherRules.MapTypeRules...)
	r.CallExprRules = append(r.CallExprRules, otherRules.CallExprRules...)
	r.ReturnStmtRules = append(r.ReturnStmtRules, otherRules.ReturnStmtRules...)

	return r
}
