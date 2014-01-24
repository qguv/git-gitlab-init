/*
Initializes a repository remotely for GitLab-hosted origin servers.
 */

package main

import (
    "os"
    "fmt"
    "github.com/docopt/docopt.go"
)

var (
    version string = "0.1.0"
    argsToParse []string = os.Args[1:]
    automaticHelp bool = true
    optionsFirst bool = true
    helpstring string = `git gitlab-init

Create an empty Git repository on GitLab and locally.

Usage:
  git gitlab-init [-u <api url>] [-k <api key>] [--] <repository> [<directory>]

Options:
  -h --help     Show this screen.
  -u            Specify GitLab api url.
  -k            Specify GitLab api key.
`
)

func main() {
    args, err := docopt.Parse(helpstring, argsToParse, automaticHelp, version, optionsFirst)
    if err != nil {
        panic(err)
    }
    fmt.Println(args)
}
