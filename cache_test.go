package pepperlint

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestCacheCurrentPackage(t *testing.T) {
	cases := []struct {
		name            string
		packages        Packages
		pkgImportPath   string
		expectedOk      bool
		expectedPackage *Package
	}{
		{
			name:       "empty package",
			packages:   Packages{},
			expectedOk: false,
		},
		{
			name: "simple package case",
			packages: Packages{
				"foo": &Package{},
			},
			pkgImportPath:   "foo",
			expectedOk:      true,
			expectedPackage: &Package{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := Cache{
				Packages:             c.packages,
				currentPkgImportPath: c.pkgImportPath,
			}

			pkg, ok := cache.CurrentPackage()

			if e, a := c.expectedOk, ok; e != a {
				t.Errorf("expected %t, but received %t", e, a)
			}

			if e, a := c.expectedPackage, pkg; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}

func TestCacheCurrentFile(t *testing.T) {
	fileAST := &ast.File{}
	cases := []struct {
		name          string
		packages      Packages
		pkgImportPath string
		fileAST       *ast.File
		expectedOk    bool
		expectedFile  *File
	}{
		{
			name:       "empty package",
			packages:   Packages{},
			expectedOk: false,
		},
		{
			name: "simple package case",
			packages: Packages{
				"foo": &Package{
					Files: Files{
						{
							ASTFile: fileAST,
						},
					},
				},
			},
			pkgImportPath: "foo",
			fileAST:       fileAST,
			expectedOk:    true,
			expectedFile: &File{
				ASTFile: fileAST,
			},
		},
		{
			name: "invalid package case",
			packages: Packages{
				"foo": &Package{
					Files: Files{
						{
							ASTFile: nil,
						},
					},
				},
			},
			pkgImportPath: "foo",
			fileAST:       fileAST,
			expectedOk:    false,
			expectedFile:  nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := Cache{
				Packages:             c.packages,
				currentPkgImportPath: c.pkgImportPath,
				currentFile:          c.fileAST,
			}

			file, ok := cache.CurrentFile()

			if e, a := c.expectedOk, ok; e != a {
				t.Errorf("expected %t, but received %t", e, a)
			}

			if e, a := c.expectedFile, file; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}

func TestCachePackagesGet(t *testing.T) {
	cases := []struct {
		name            string
		packages        Packages
		importPath      string
		expectedOk      bool
		expectedPackage *Package
	}{
		{
			name:       "empty package",
			packages:   Packages{},
			expectedOk: false,
		},
		{
			name: "found package",
			packages: Packages{
				"bar": &Package{
					Files: Files{
						{},
					},
				},
			},
			importPath: "bar",
			expectedOk: true,
			expectedPackage: &Package{
				Files: Files{
					{},
				},
			},
		},
		{
			name: "invalid package import path",
			packages: Packages{
				"bar": &Package{
					Files: Files{
						{},
					},
				},
			},
			importPath: "baz",
			expectedOk: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := Cache{
				Packages: c.packages,
			}

			pkg, ok := cache.Packages.Get(c.importPath)

			if e, a := c.expectedOk, ok; e != a {
				t.Errorf("expected %t, but received %t", e, a)
			}

			if e, a := c.expectedPackage, pkg; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}

func TestCacheGetTypeInfo(t *testing.T) {
	cases := []struct {
		name             string
		packages         Packages
		pkgImportPath    string
		typeName         string
		expectedOk       bool
		expectedTypeInfo TypeInfo
	}{
		{
			name: "empty type name case",
			packages: Packages{
				"bar": &Package{
					Files: Files{
						{},
					},
				},
			},
			pkgImportPath: "bar",
		},
		{
			name: "valid type info",
			packages: Packages{
				"bar": &Package{
					Files: Files{
						{
							TypeInfos: TypeInfos{
								"name": {
									PkgName: "baz",
								},
							},
						},
					},
				},
			},
			typeName:      "name",
			pkgImportPath: "bar",
			expectedOk:    true,
			expectedTypeInfo: TypeInfo{
				PkgName: "baz",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := Cache{
				Packages:             c.packages,
				currentPkgImportPath: c.pkgImportPath,
			}

			pkg, ok := cache.CurrentPackage()
			if !ok {
				t.Errorf("unexpected missing package")
			}

			info, ok := pkg.Files.GetTypeInfo(c.typeName)
			if e, a := c.expectedOk, ok; e != a {
				t.Errorf("expected %t, but received %t", e, a)
			}

			if e, a := c.expectedTypeInfo, info; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}

func TestCacheGetOpInfo(t *testing.T) {
	cases := []struct {
		name           string
		packages       Packages
		pkgImportPath  string
		opName         string
		expectedOk     bool
		expectedOpInfo OpInfo
	}{
		{
			name: "empty type name case",
			packages: Packages{
				"bar": &Package{
					Files: Files{
						{},
					},
				},
			},
			pkgImportPath: "bar",
		},
		{
			name: "valid op info",
			packages: Packages{
				"bar": &Package{
					Files: Files{
						{
							OpInfos: OpInfos{
								"name": {
									PkgName: "baz",
								},
							},
						},
					},
				},
			},
			opName:        "name",
			pkgImportPath: "bar",
			expectedOk:    true,
			expectedOpInfo: OpInfo{
				PkgName: "baz",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cache := Cache{
				Packages:             c.packages,
				currentPkgImportPath: c.pkgImportPath,
			}

			pkg, ok := cache.CurrentPackage()
			if !ok {
				t.Errorf("unexpected missing package")
			}

			info, ok := pkg.Files.GetOpInfo(c.opName)
			if e, a := c.expectedOk, ok; e != a {
				t.Errorf("expected %t, but received %t", e, a)
			}

			if e, a := c.expectedOpInfo, info; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}

func TestCacheVisitorImports(t *testing.T) {
	cases := []struct {
		name            string
		code            string
		expectedImports map[string]string
	}{
		{
			name: "import specs",
			code: `package foo
			import (
				"bar/baz"
			)

			import "moo/cow"
			`,
			expectedImports: map[string]string{
				"baz": "bar/baz",
				"cow": "moo/cow",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fset := token.NewFileSet()
			node, err := parser.ParseFile(fset, c.name, c.code, parser.ParseComments)
			if err != nil {
				t.Errorf("unexpected error %v", err)
			}

			if node == nil {
				t.Fatal("unexpected nil expr")
			}

			cache := &Cache{
				Packages: Packages{
					"": &Package{},
				},
			}
			ast.Walk(cache, node)

			pkg, ok := cache.Packages.Get("")
			if !ok {
				t.Errorf("expected package to be defined")
			}

			if len(pkg.Files) != 1 {
				t.Errorf("expected 1 file, but received %d", len(pkg.Files))
			}

			f := pkg.Files[0]

			if e, a := c.expectedImports, f.Imports; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}

func TestOpInfoMethod(t *testing.T) {
	spec := &ast.TypeSpec{}

	cases := []struct {
		name     string
		opInfo   OpInfo
		spec     *ast.TypeSpec
		expected bool
	}{
		{
			name: "is function",
			opInfo: OpInfo{
				TypeSpecs: []*ast.TypeSpec{
					spec,
					spec,
					spec,
				},
			},
		},
		{
			name: "is method but invalid spec",
			opInfo: OpInfo{
				IsMethod: true,
				TypeSpecs: []*ast.TypeSpec{
					spec,
				},
			},
			expected: false,
		},
		{
			name: "is method",
			opInfo: OpInfo{
				IsMethod: true,
				TypeSpecs: []*ast.TypeSpec{
					spec,
				},
			},
			spec:     spec,
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if e, a := c.opInfo.HasReceiverType(c.spec), c.expected; e != a {
				t.Errorf("expected %t, but received %t", e, a)
			}
		})
	}
}
