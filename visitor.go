package pepperlint

import (
	"go/ast"
	"go/token"
	"path/filepath"
)

// Visitor is used to traferse a node and run the proper validaters
// based on the node that is passed in.
type Visitor struct {
	Rules  Rules
	Errors Errors

	PackagesCache *Cache

	FileSet *token.FileSet

	currentPkgImportPath string
}

// NewVisitor returns a new visitor and instantiates a new rule set from
// the adders parameter.
func NewVisitor(fset *token.FileSet, cache *Cache, opts ...Option) *Visitor {
	v := &Visitor{
		FileSet:       fset,
		PackagesCache: cache,
	}

	for _, o := range opts {
		if opt, ok := o.(FileSetOption); ok {
			opt.WithFileSet(fset)
		}

		if opt, ok := o.(RulesAdder); ok {
			opt.AddRules(&v.Rules)
		}

		if opt, ok := o.(CacheOption); ok {
			opt.WithCache(v.PackagesCache)
		}
	}

	return v
}

// Visit is our generic visitor that will visit each ast type and call
// the appropriate rules based on what type the node is.
func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return v
	}

	//Log("VISITING %p %T %v", node, node, node)

	switch t := node.(type) {
	case *ast.Package:
		for k := range t.Files {
			v.PackagesCache.CurrentPkgImportPath = GetImportPathFromFullPath(filepath.Dir(k))
			break
		}

		v.visitPackage(t)
	case *ast.File:
		v.PackagesCache.CurrentASTFile = t

		v.visitFile(t)
	case ast.Decl:
		v.visitDecl(t)
	case ast.Expr:
		// ignored due to visiting of ExprStmt in
		// visitStmt
		if binExpr, ok := t.(*ast.BinaryExpr); ok {
			v.visitBinaryExpr(binExpr)
		}
	case ast.Spec:
		v.visitSpec(t)
	case ast.Stmt:
		v.visitStmt(t)
		// TOOD: May contain a bug that visits twice for both visitField
		// and visitFieldList
	case *ast.Field:
		v.visitField(t)
	case *ast.FieldList:
		v.visitFieldList(t)
	case *ast.Comment:
	case *ast.CommentGroup:
	default:
		if t != nil {
			Log("TODO: visit %T\n", t)
		}
		return v
	}

	return v
}

func (v *Visitor) visitDecl(decl ast.Decl) {
	switch t := decl.(type) {
	case *ast.BadDecl:
		// TODO: add error of bad declaration here
		Log("TODO: visit %T\n", t)
	case *ast.FuncDecl:
		v.visitFuncDecl(t)
	case *ast.GenDecl:
		v.visitGenDecl(t)
	default:
		Log("TODO: visitDecl %T\n", t)
	}
}

func (v *Visitor) visitExpr(expr ast.Expr) {
	switch t := expr.(type) {
	// covered by visitSpec which is why these do nothing
	case *ast.Ident:
	case *ast.ParenExpr:
	case *ast.SelectorExpr:
	case *ast.StarExpr:
	case *ast.ArrayType:
	case *ast.ChanType:
	case *ast.FuncType:
		v.visitFuncType(t)
	case *ast.InterfaceType:
	case *ast.MapType:
	case *ast.StructType:
	// Not covered by visitSpec
	case *ast.BasicLit:
		Log("TODO: visitExpr %T\n", t)
	case *ast.CompositeLit:
		Log("TODO: visitExpr %T\n", t)
	case *ast.CallExpr:
		v.visitCallExpr(t)
	case *ast.BinaryExpr:
		v.visitBinaryExpr(t)
	case *ast.IndexExpr:
		Log("TODO: visitExpr %T\n", t)
	case *ast.KeyValueExpr:
		Log("TODO: visitExpr %T\n", t)
	default:
		Log("T %T\n", t)
	}
}

func (v *Visitor) visitSpec(spec ast.Spec) {
	switch t := spec.(type) {
	case *ast.ImportSpec:
		//log.Println("TODO: visit *ast.ImportSpec")
	case *ast.TypeSpec:
		v.visitTypeSpec(t)
	case *ast.ValueSpec:
		v.visitValueSpec(t)
	default:
		Log("TODO: visitDecl %T\n", t)
	}
}

func (v *Visitor) visitStmt(stmt ast.Stmt) {
	switch t := stmt.(type) {
	case *ast.AssignStmt:
		v.visitAssignStmt(t)
	case *ast.BlockStmt:
		v.visitBlockStmt(t)
	case *ast.ExprStmt:
		v.visitExpr(t.X)
	case *ast.ReturnStmt:
		v.visitReturnStmt(t)
	case *ast.IncDecStmt:
		v.visitIncDecStmt(t)
	case *ast.RangeStmt:
		v.visitRangeStmt(t)
	default:
		Log("TODO: visitStmt %T\n", t)
	}
}

func (v *Visitor) visitTypeSpec(spec *ast.TypeSpec) {
	if err := v.Rules.TypeSpecRules.ValidateTypeSpec(spec); err != nil {
		v.Errors.Add(err)
	}

	switch t := spec.Type.(type) {
	case *ast.Ident:
		Log("TODO: visit *ast.Ident")
	case *ast.ParenExpr:
		Log("TODO: visit *ast.ParenExpr")
	case *ast.SelectorExpr:
		Log("TODO: visit *ast.SelectorExpr")
	case *ast.StarExpr:
		Log("TODO: visit *ast.StarExpr")

	// Types
	case *ast.ArrayType:
		v.visitArrayType(t)
	case *ast.ChanType:
		v.visitChanType(t)
	case *ast.FuncType:
		v.visitFuncType(t)
	case *ast.InterfaceType:
		v.visitInterfaceType(t)
	case *ast.MapType:
		v.visitMapType(t)
	case *ast.StructType:
		v.visitStructType(t)
	default:
		Log("TODO: visitType %T\n", t)
	}
}

func (v *Visitor) visitValueSpec(spec *ast.ValueSpec) {
	if err := v.Rules.ValueSpecRules.ValidateValueSpec(spec); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitStructType(s *ast.StructType) {
	if err := v.Rules.StructTypeRules.ValidateStructType(s); err != nil {
		v.Errors.Add(err)
	}

	v.visitFieldList(s.Fields)
}

func (v *Visitor) visitFieldList(fields *ast.FieldList) {
	if err := v.Rules.FieldListRules.ValidateFieldList(fields); err != nil {
		v.Errors.Add(err)
	}

	for _, field := range fields.List {
		v.visitField(field)
	}
}

func (v *Visitor) visitField(field *ast.Field) {
	if err := v.Rules.FieldRules.ValidateField(field); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitArrayType(array *ast.ArrayType) {
	if err := v.Rules.ArrayTypeRules.ValidateArrayType(array); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitMapType(m *ast.MapType) {
	if err := v.Rules.MapTypeRules.ValidateMapType(m); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitChanType(ch *ast.ChanType) {
	if err := v.Rules.ChanTypeRules.ValidateChanType(ch); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitFuncType(fn *ast.FuncType) {
	if err := v.Rules.FuncTypeRules.ValidateFuncType(fn); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitInterfaceType(iface *ast.InterfaceType) {
	if err := v.Rules.InterfaceTypeRules.ValidateInterfaceType(iface); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitFuncDecl(fnDecl *ast.FuncDecl) {
	if err := v.Rules.FuncDeclRules.ValidateFuncDecl(fnDecl); err != nil {
		v.Errors.Add(err)
	}
}

// visitGenDecl will happen before any visiting of more specific specs, ie XXXSpec.
// This function can be used to grab documentation or other metadata to further
// validation used on more specific rules.
//
// An example of this would be how DeprecateStructRule works. The rule will visit
// general declarations first to populate the documentation of the type spec. Docs
// need to be pulled from GenDecl due to the semantic meaning of `type name struct`.
//
// type foo struct{} is a shortcut for
//
// // GenDecl docs
// type (
//     // TypeSpec docs!
//     foo struct{}
// )
func (v *Visitor) visitGenDecl(decl *ast.GenDecl) *Visitor {
	if err := v.Rules.GenDeclRules.ValidateGenDecl(decl); err != nil {
		v.Errors.Add(err)
	}

	return v
}

func (v *Visitor) visitAssignStmt(stmt *ast.AssignStmt) {
	if err := v.Rules.AssignStmtRules.ValidateAssignStmt(stmt); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitBlockStmt(stmt *ast.BlockStmt) {
	if err := v.Rules.BlockStmtRules.ValidateBlockStmt(stmt); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitCallExpr(expr *ast.CallExpr) {
	if err := v.Rules.CallExprRules.ValidateCallExpr(expr); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitReturnStmt(stmt *ast.ReturnStmt) {
	if err := v.Rules.ReturnStmtRules.ValidateReturnStmt(stmt); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitFile(f *ast.File) {
	if err := v.Rules.FileRules.ValidateFile(f); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitPackage(pkg *ast.Package) {
	if err := v.Rules.PackageRules.ValidatePackage(pkg); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitBinaryExpr(expr *ast.BinaryExpr) {
	if err := v.Rules.BinaryExprRules.ValidateBinaryExpr(expr); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitIncDecStmt(stmt *ast.IncDecStmt) {
	if err := v.Rules.IncDecStmtRules.ValidateIncDecStmt(stmt); err != nil {
		v.Errors.Add(err)
	}
}

func (v *Visitor) visitRangeStmt(stmt *ast.RangeStmt) {
	if err := v.Rules.RangeStmtRules.ValidateRangeStmt(stmt); err != nil {
		v.Errors.Add(err)
	}
}
