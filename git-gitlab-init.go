/*
Initializes a repository remotely for GitLab-hosted origin servers.
*/

package main

import (
    "fmt"
    "github.com/docopt/docopt.go"
    "os"
)

var (
    // Variables about this program
    version       string   = "0.1.0"
    argsToParse   []string = os.Args[1:]
    automaticHelp bool     = true
    optionsFirst  bool     = true
    helpstring    string   = `git gitlab-init

Create an empty Git repository on GitLab and locally.

The commands below can be used as "git gitlab-init" or as "git-gitlab-init".

Usage:
  git-gitlab-init (-h | --help | --version)
  git-gitlab-init [-u API_URL] [-k API_KEY] [--] <repository> [<directory>]

Options:
  -h, --help    Show this screen and exit.
  -u API_URL    Specify GitLab api url.
  -k API_KEY    Specify GitLab api key.
  --version     Print version and exit.
`
)

func main() {
    args, err := docopt.Parse(helpstring, argsToParse, automaticHelp, version, optionsFirst)
    if err != nil {
        panic(err)
    }
    fmt.Println(args)
}
