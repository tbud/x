package main

// import ()

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

func newCommand(cmd *Command, args []string) {

}
