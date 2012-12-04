package c

import (
	"go/parser"
)

// CompileDep returns the names of all packages required to build pkgs
// that are inside the gopath tree.
func (b *Build) CompileDep(pkgs ...string) ([]string, error) {
	depTabl := newDepTable(b)
	for _, pkg := range pkgs {
		if err := depTabl.Add(pkg); err != nil {
			return nil, err
		}
	}
	return depTabl.All(), nil
}

// depTable maintains the dependent packages for a list of incrementally added
// target packages
type depTable struct {
	build  *Build
	pkgs   map[string]*depPkg
	follow []string
}

type depPkg struct {
	imports  []string
}

func newDepTable(b *Build) *depTable {
	return &depTable{
		build:  b,
		pkgs:   make(map[string]*depPkg),
		follow: nil,
	}
}

func (dt *depTable) Add(pkg string) error {
	dt.follow = append(dt.follow, pkg)
	return dt.loop()
}

func (dt *depTable) loop() error {
	for len(dt.follow) > 0 {
		pop := dt.follow[0]
		dt.follow = dt.follow[1:]

		// Check if package already processed
		if _, present := dt.pkgs[pop]; present {
			continue
		}

		// Parse package source
		skel, err := dt.build.ParsePkg(pop, parser.ImportsOnly)
		if err != nil {
			return err
		}

		// Process all import specs in all source files
		imps := make(map[string]struct{})
		for _, pkg := range skel.Pkgs {
			pimps := pkgImports(pkg)
			for i, _ := range pimps {
				imps[i] = struct{}{}
			}
		}

		// Make pkg structure
		dpkg := &depPkg{}
		for pkg, _ := range imps {
			dpkg.imports = append(dpkg.imports, pkg)
		}

		// Save pkg structure
		dt.pkgs[pop] = dpkg
	}
	return nil
}

func (dt *depTable) All() []string {
	var all []string
	for pkg, _ := range dt.pkgs {
		all = append(all, pkg)
	}
	return all
}