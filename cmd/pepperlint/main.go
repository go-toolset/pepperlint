package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"go/ast"
	"go/parser"
	"go/token"

	"github.com/go-toolset/pepperlint"
)

// PackageSetBuilder wil lint the directory specified and walk the directory to grab
// the necessary dependencies needed for the given rules.
type PackageSetBuilder struct {
	pkg         string
	includeDirs []string
}

// Packages contains all packages in a given path
type Packages struct {
	path string
	pkgs map[string]*ast.Package
}

// Container contains two separate Package groups. 'RulesPackage' represents packages
// that rules are applied to and while the other, 'Packages', will have rules applied
// during visitation.
type Container struct {
	RulesPackages []Packages

	// Packages do not have rules applied to them
	Packages []Packages
}

// WithPkg will return a copy of the builder with the dir being
// the root directory that will be walked during linting
func (b PackageSetBuilder) WithPkg(pkg string) PackageSetBuilder {
	b.pkg = pkg
	return b
}

// WithImports will return a copy of the builder with the imports to be
// required for determining what the rules are acting against.
func (b PackageSetBuilder) WithImports(imports []string) PackageSetBuilder {
	b.includeDirs = imports
	return b
}

func (b PackageSetBuilder) addDir(fset *token.FileSet, container Container) Container {
	// walk root directory to gather all packages in the given directory
	filepath.Walk(b.pkg, func(path string, info os.FileInfo, err error) error {
		// only need directories due to the `ParseDir` call.
		if !info.Mode().IsDir() {
			return nil
		}

		root, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
		if err != nil {
			log.Fatal(err)
		}

		container.RulesPackages = append(container.RulesPackages, Packages{
			path: path,
			pkgs: root,
		})

		return nil
	})

	return container
}

func (b PackageSetBuilder) addFile(fset *token.FileSet, container Container) (Container, error) {
	root, err := parser.ParseFile(fset, b.pkg, nil, parser.ParseComments)
	if err != nil {
		return container, err
	}

	pkgRoot := map[string]*ast.Package{
		root.Name.Name: &ast.Package{
			Name: root.Name.Name,
			Files: map[string]*ast.File{
				filepath.Base(b.pkg): root,
			},
		},
	}

	container.RulesPackages = append(container.RulesPackages, Packages{
		path: b.pkg,
		pkgs: pkgRoot,
	})

	return container, nil
}

// Build will build a container which contains all packages we will walk.
func (b PackageSetBuilder) Build() (Container, *token.FileSet, error) {
	container := Container{}
	fset := token.NewFileSet()

	info, err := os.Stat(b.pkg)
	if err != nil {
		return container, nil, err
	}

	if info.IsDir() {
		container = b.addDir(fset, container)
	} else {
		container, err = b.addFile(fset, container)
		if err != nil {
			return container, nil, err
		}
	}

	for _, included := range b.includeDirs {
		filepath.Walk(included, func(path string, info os.FileInfo, err error) error {
			if !info.Mode().IsDir() {
				return nil
			}

			root, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}

			container.Packages = append(container.Packages, Packages{
				path: path,
				pkgs: root,
			})
			return nil
		})
	}

	return container, fset, nil
}

func walk(v ast.Visitor, p []Packages) {
	sortedPkgNames := []string{}
	for _, pkgs := range p {
		sortedPkgNames = sortedPkgNames[0:0]
		for name := range pkgs.pkgs {
			sortedPkgNames = append(sortedPkgNames, name)
		}

		sort.Strings(sortedPkgNames)
		for _, name := range sortedPkgNames {
			pkg := pkgs.pkgs[name]
			ast.Walk(v, pkg)
		}
	}
}

// lint will lint the dir while walking the dirs provided to grab necessary metadata
// from to then validate the dir with the gathered metadata.
func lint(config Config, pkgs []string, pkg string) (*pepperlint.Visitor, Container, error) {
	// Prepends go path
	gopath := filepath.Join(os.Getenv("GOPATH"), "src")
	pkg = filepath.Join(gopath, pkg)
	for i, p := range pkgs {
		pkgs[i] = filepath.Join(gopath, p)
	}

	builder := PackageSetBuilder{}.WithImports(pkgs).WithPkg(pkg)

	container, fset, err := builder.Build()
	if err != nil {
		return nil, Container{}, err
	}

	cache := pepperlint.NewCache()

	//rule := core.NewDeprecatedRule(fset)
	v := pepperlint.NewVisitor(fset, cache, config.Options()...)

	walk(cache, container.Packages)
	walk(cache, container.RulesPackages)
	walk(v, container.RulesPackages)

	return v, container, nil
}

func suppress(suppressions Suppressions, errs pepperlint.Errors) []error {
	s := map[string]File{}
	validErrs := []error{}

	for _, sup := range suppressions {
		if sup.File == nil {
			continue
		}

		s[sup.File.FilePath] = *sup.File
	}

	for _, err := range errs {
		e, ok := err.(pepperlint.FileError)
		if !ok {
			validErrs = append(validErrs, err)
			continue
		}

		supRule, ok := s[e.Filename()]
		if !ok {
			validErrs = append(validErrs, err)
			continue
		}

		if supRule.LineNumber == nil {
			continue
		}

		if e.LineNumber() == *supRule.LineNumber {
			continue
		}

		validErrs = append(validErrs, err)
	}

	return validErrs
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("directory needs to be provided")
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	f := newFlags()
	config := buildConfig(f.ConfigPath)
	config = f.Merge(config)

	// TODO:
	// Do we still need to move this into the pkgs variable?
	// Can we not use config.IncludePkgs instead?
	pkgs := []string{}
	for _, p := range config.IncludePkgs {
		if len(p) == 0 {
			continue
		}

		pkgs = append(pkgs, p)
	}

	pkg := os.Args[len(os.Args)-1]
	v, _, err := lint(config, pkgs, pkg)
	if err != nil {
		panic(err)
	}

	errs := suppress(config.Suppressions, v.Errors)
	if len(errs) != 0 {
		fmt.Fprintf(os.Stderr, "%v", errs)
		os.Exit(1)
	}
}
