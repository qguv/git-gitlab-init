/*
Initializes a repository remotely for GitLab-hosted origin servers.
*/

package main

import (
    "fmt"
    "github.com/docopt/docopt.go"
    "io/ioutil"
    "net/http"
    "os"
    "os/exec"
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
  git-gitlab-init [-u API_VERSION] [-k API_KEY] [--] <repository> [<directory>]

Options:
  -h, --help        Show this screen and exit.
  -u API_VERSION    Specify GitLab api version url [default: v3].
  -k API_KEY        Specify GitLab api key.
  --version         Print version and exit.
`
)

func runCommand(name string, arg ...string) (string, error) {
    cmd := exec.Command(name, arg...)
    raw_out, err := cmd.Output()
    out := string(raw_out)
    return out, err
}

func getSetting(name string) string {
    cmd := exec.Command("git", "config", "--get", name)
    raw_out, err := cmd.Output()
    if err != nil {
        panic(err)
    }
    return string(raw_out)
}

func sendGet(s string) string {
    fmt.Println(s) //debug
    resp, err := http.Get("http://example.com/")
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    return string(body)
}

/*
   git init
   touch README
   git add README
   git commit -m 'first commit'
   git remote add origin http://src.nascifi.com:8080/quintus/tmp-testing.git
   git push -u origin master
*/

func complainUndefined(options [][2]string) {
    fmt.Println("Error! Your Gitlab API settings aren't defined.")
    fmt.Println("Try running the following:\n")
    var param [2]string
    for _, param = range options {
        setting, desc := param[0], param[1]
        fmt.Println("    git config --global " + setting + " \"" + desc + "\"")
    }
    fmt.Println("")
}

func initialize(projectName string, user string, url string, dir string) (ok bool) {
    runCommand("git", "init")

    f, err := os.OpenFile("README.md", os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    _, err = f.WriteString("# " + projectName)
    if err != nil {
        panic(err)
    }

    runCommand("git", "add", "README.md")
    runCommand("git", "commit", "-m", "initial commit")
    runCommand("git", "remote", "add", "origin", "")
    return true // ok
}

func main() {
    args, err := docopt.Parse(helpstring, argsToParse, automaticHelp, version, optionsFirst)
    if err != nil {
        panic(err)
    }
    var options [][2]string
    options = append(options, [2]string{"user.name", "First Last"})
    options = append(options, [2]string{"gitlab.api", "v3"})
    complainUndefined(options)
}
