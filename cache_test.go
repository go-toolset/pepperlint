package pepperlint

import (
	"go/ast"
	"reflect"
	"testing"
)

func TestCacheCurrentPackage(t *testing.T) {
	cases := []struct {
		Name            string
		Packages        Packages
		PkgImportPath   string
		expectedOk      bool
		expectedPackage *Package
	}{
		{
			Name:       "empty package",
			Packages:   Packages{},
			expectedOk: false,
		},
		{
			Name: "simple package case",
			Packages: Packages{
				"foo": &Package{},
			},
			PkgImportPath:   "foo",
			expectedOk:      true,
			expectedPackage: &Package{},
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			cache := Cache{
				Packages:             c.Packages,
				currentPkgImportPath: c.PkgImportPath,
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
		Name          string
		Packages      Packages
		PkgImportPath string
		FileAST       *ast.File
		expectedOk    bool
		expectedFile  *File
	}{
		{
			Name:       "empty package",
			Packages:   Packages{},
			expectedOk: false,
		},
		{
			Name: "simple package case",
			Packages: Packages{
				"foo": &Package{
					Files: Files{
						{
							ASTFile: fileAST,
						},
					},
				},
			},
			PkgImportPath: "foo",
			FileAST:       fileAST,
			expectedOk:    true,
			expectedFile: &File{
				ASTFile: fileAST,
			},
		},
		{
			Name: "invalid package case",
			Packages: Packages{
				"foo": &Package{
					Files: Files{
						{
							ASTFile: nil,
						},
					},
				},
			},
			PkgImportPath: "foo",
			FileAST:       fileAST,
			expectedOk:    false,
			expectedFile:  nil,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			cache := Cache{
				Packages:             c.Packages,
				currentPkgImportPath: c.PkgImportPath,
				currentFile:          c.FileAST,
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
		Name            string
		Packages        Packages
		ImportPath      string
		expectedOk      bool
		expectedPackage *Package
	}{
		{
			Name:       "empty package",
			Packages:   Packages{},
			expectedOk: false,
		},
		{
			Name: "found package",
			Packages: Packages{
				"bar": &Package{
					Files: Files{
						{},
					},
				},
			},
			ImportPath: "bar",
			expectedOk: true,
			expectedPackage: &Package{
				Files: Files{
					{},
				},
			},
		},
		{
			Name: "invalid package import path",
			Packages: Packages{
				"bar": &Package{
					Files: Files{
						{},
					},
				},
			},
			ImportPath: "baz",
			expectedOk: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			cache := Cache{
				Packages: c.Packages,
			}

			pkg, ok := cache.Packages.Get(c.ImportPath)

			if e, a := c.expectedOk, ok; e != a {
				t.Errorf("expected %t, but received %t", e, a)
			}

			if e, a := c.expectedPackage, pkg; !reflect.DeepEqual(e, a) {
				t.Errorf("expected %v, but received %v", e, a)
			}
		})
	}
}
