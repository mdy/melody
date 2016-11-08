package extract

import (
	"github.com/mdy/melody/project"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

// Extract package configuration from its directory
func ProjectConfig(dir string) (*project.Config, error) {
	pkg, err := build.Default.ImportDir(dir, build.FindOnly)
	if err != nil {
		return nil, err
	}

	config := &project.Config{Version: "0.1.0"}
	config.Name = pkg.ImportPath

	deps, err := extractDependencies(config.Name, dir)
	if err != err {
		return config, err
	}

	config.Dependencies = deps
	return config, nil
}

// Extract package information from repository dir
// Adopted from "matchPackagesInFS" in "cmd/go"
func extractDependencies(pkgName, dir string) (map[string]string, error) {
	deps := map[string]string{}
	context := build.Default

	// Walk each project directory and run build.Context.ImportDir
	filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() {
			return nil
		}

		// filepath.Walk starts at dir and recurses. For the recursive case,
		// the path is the result of filepath.Join, which calls filepath.Clean.
		// The initial case is not Cleaned, though, so we do this explicitly.
		if path == dir {
			path = filepath.Clean(path)
		}

		_, elem := filepath.Split(path)
		dot := strings.HasPrefix(elem, ".") && elem != "." && elem != ".."
		if dot || strings.HasPrefix(elem, "_") || elem == "testdata" || elem == "vendor" {
			return filepath.SkipDir
		}

		subPkg, _ := filepath.Rel(dir, path)
		pkg, err := context.Import(pkgName+"/"+subPkg, dir, build.IgnoreVendor)
		if err != nil {
			return nil
		}

		for _, d := range append(pkg.Imports, pkg.TestImports...) {
			if !strings.HasPrefix(d, pkgName+"/") && !isStandardImportPath(d) {
				deps[d] = "head"
			}
		}

		return nil
	})

	return deps, nil
}

// From https://github.com/golang/go/blob/master/src/cmd/go/pkg.go
func isStandardImportPath(path string) bool {
	i := strings.Index(path, "/")
	if i < 0 {
		i = len(path)
	}
	return !strings.Contains(path[:i], ".")
}
