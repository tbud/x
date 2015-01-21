package main

import (
	"bytes"
	"github.com/tbud/bud"
	"go/build"
	"os/exec"
	"path/filepath"
)

var cmdNew = &Command{
	UsageLine: "new [path] [skeleton]",
	Short:     "create a skeleton Bud application",
	Long: `
New creates a few files to get a new bud application running quickly.

It puts all of the files in the given import path, taking the final element in
the path to be the app name.

Skeleton is an optional argument, provided as an import path

For example:

    bud new import/path/helloworld

    bud new import/path/helloworld import/path/skeleton
	`,
}

func init() {
	cmdNew.Run = newCommand
}

var (
	// go related paths
	gopath  string
	gocmd   string
	srcRoot string

	// bud related paths
	budPkg       *build.Package
	appPath      string
	appName      string
	basePath     string
	importPath   string
	skeletonPath string
)

func newCommand(cmd *Command, args []string) {
	if len(args) == 0 {
		fatalf("No import path given.\nRun 'bud help new' for usage.\n")
	}
	if len(args) > 2 {
		fatalf("Too many arguments provided.\nRun 'bud help new' for usage.\n")
	}

	// checking and setting go paths
	initGoPaths()

	// checking and setting application
	setApplicationPath(args)

	// checking and setting skeleton
	setSkeletonPath(args)

	// copy files to new app directory
	copyNewAppFiles()

	logf("Your application is ready:\n  %s", appPath)
	logf("\nYou can run it with:\n  bud run %s", importPath)
}

// lookup and set Go related variables
func initGoPaths() {
	// lookup go path
	gopath = build.Default.GOPATH
	if gopath == "" {
		fatalf("Abort: GOPATH environment variable is not set. " +
			"Please refer to http://golang.org/doc/code.html to configure your Go environment.")
	}

	// set go src path
	srcRoot = filepath.Join(filepath.SplitList(gopath)[0], "src")

	// check for go executable
	var err error
	gocmd, err = exec.LookPath("go")
	if err != nil {
		fatalf("Go executable not found in PATH.")
	}
}

func setApplicationPath(args []string) {
	var err error
	importPath = args[0]
	if filepath.IsAbs(importPath) {
		fatalf("Abort: '%s' looks like a directory.  Please provide a Go import path instead.",
			importPath)
	}

	_, err = build.Import(importPath, "", build.FindOnly)
	if err == nil {
		fatalf("Abort: Import path %s already exists.\n", importPath)
	}

	budPkg, err = build.Import(bud.BUD_IMPORT_PATH, "", build.FindOnly)
	if err != nil {
		fatalf("Abort: Could not find bud source code: %s\n", err)
	}

	appPath = filepath.Join(srcRoot, filepath.FromSlash(importPath))
	appName = filepath.Base(appPath)
	basePath = filepath.ToSlash(filepath.Dir(importPath))

	if basePath == "." {
		// we need to remove the a single '.' when
		// the app is in the $GOROOT/src directory
		basePath = ""
	} else {
		// we need to append a '/' when the app is
		// is a subdirectory such as $GOROOT/src/path/to/revelapp
		basePath += "/"
	}
}

func setSkeletonPath(args []string) {
	var err error
	if len(args) == 2 { // user specified
		skeletonName := args[1]
		_, err = build.Import(skeletonName, "", build.FindOnly)
		if err != nil {
			// Execute "go get <pkg>"
			getCmd := exec.Command(gocmd, "get", "-d", skeletonName)
			logf("Exec: %s", getCmd.Args)
			bOutput, err := getCmd.CombinedOutput()

			bpos := bytes.Index(bOutput, []byte("no buildable Go source files in"))
			if err != nil && bpos == -1 {
				fatalf("Abort: Could not find or 'go get' Skeleton  source code: %s\n%s\n", bOutput, skeletonName)
			}
		}

		skeletonPath = filepath.Join(srcRoot, skeletonName)
	} else {
		skeletonPath = filepath.Join(budPkg.Dir, "skeleton")
	}
}

func copyNewAppFiles() {

}
