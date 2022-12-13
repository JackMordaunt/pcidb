package pcidb

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Concrete merged set of configuration switches that get passed to pcidb
// internal functions
type context struct {
	chroot             string
	cacheOnly          bool
	cachePath          string
	path               string
	enableNetworkFetch bool
	searchPaths        []string
}

func contextFromOptions(merged *WithOption) *context {
	ctx := &context{
		chroot:             *merged.Chroot,
		cacheOnly:          *merged.CacheOnly,
		cachePath:          getCachePath(),
		enableNetworkFetch: *merged.EnableNetworkFetch,
		path:               *merged.Path,
		searchPaths:        make([]string, 0),
	}
	ctx.setSearchPaths()
	return ctx
}

func getCachePath() string {
	hdir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed getting homedir.Dir(): %v", err)
		return ""
	}
	fp, err := Expand(filepath.Join(hdir, ".cache", "pci.ids"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed expanding local cache path: %v", err)
		return ""
	}
	return fp
}

// Depending on the operating system, sets the context's searchPaths to a set
// of local filepaths to search for a pci.ids database file
func (ctx *context) setSearchPaths() {
	// Look in direct path first, if set
	if ctx.path != "" {
		ctx.searchPaths = append(ctx.searchPaths, ctx.path)
		return
	}
	// A set of filepaths we will first try to search for the pci-ids DB file
	// on the local machine. If we fail to find one, we'll try pulling the
	// latest pci-ids file from the network
	ctx.searchPaths = append(ctx.searchPaths, ctx.cachePath)
	if ctx.cacheOnly {
		return
	}

	rootPath := ctx.chroot

	if runtime.GOOS != "windows" {
		ctx.searchPaths = append(
			ctx.searchPaths,
			filepath.Join(rootPath, "usr", "share", "hwdata", "pci.ids"),
		)
		ctx.searchPaths = append(
			ctx.searchPaths,
			filepath.Join(rootPath, "usr", "share", "misc", "pci.ids"),
		)
		ctx.searchPaths = append(
			ctx.searchPaths,
			filepath.Join(rootPath, "usr", "share", "hwdata", "pci.ids.gz"),
		)
		ctx.searchPaths = append(
			ctx.searchPaths,
			filepath.Join(rootPath, "usr", "share", "misc", "pci.ids.gz"),
		)
	}
}

// Expand expands the path to include the home directory if the path
// is prefixed with `~`. If it isn't prefixed with `~`, the path is
// returned as-is.
func Expand(path string) (string, error) {
	if len(path) == 0 {
		return path, nil
	}

	if path[0] != '~' {
		return path, nil
	}

	if len(path) > 1 && path[1] != '/' && path[1] != '\\' {
		return "", fmt.Errorf("cannot expand user-specific home dir")
	}

	dir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, path[1:]), nil
}
