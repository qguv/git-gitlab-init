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
    "strings"
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
  git-gitlab-init [-p PRIVACYLEVEL] [-u USERNAME] [-l URL] [-d DESCRIPTION]
                  [-v API_VERSION] [-t API_TOKEN] [--] <repository> [<directory>]

Arguments:
  <repository>      Specify repository name.
  <directory>       Optionally specify local directory name.

Options:
  -h, --help        Shows this screen and exit.
  -p PRIVACYLEVEL   Sets viewing permission status of repository. Valid options
                    are public, private, or internal. [default: private]
  -d DESCRIPTION    Specify repository description.
  -u USERNAME       Specify Gitlab username.
  -l URL            Specify Gitlab instance url.
  -v API_VERSION    Specify Gitlab api version url. [default: v3]
  -t API_TOKEN      Specify Gitlab api token, found in your Gitlab profile settings.
  --version         Prints version and exits.
`
)

func findInSlice(s string, slice [][2]string) int {
    for i, v := range slice {
        if (v[0] == s) {
            return i
        }
    }
    return -1
}

func removeGood(s string, slice [][2]string) [][2]string {
    i := findInSlice(s, slice)
    if i != -1 {
        return append(slice[:i], slice[i+1:]...)
    }
    return [][2]string{}
}

func runCommand(name string, arg ...string) (string, error) {
    cmd := exec.Command(name, arg...)
    raw_out, err := cmd.Output()
    out := string(raw_out)
    return out, err
}

/*
   git init
   touch README
   git add README
   git commit -m 'first commit'
   git remote add origin http://src.nascifi.com:8080/quintus/tmp-testing.git
   git push -u origin master
*/

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

func sendGet(url string) string {
    resp, err := http.Get(url)
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

func apiCommand(url string, data string, token string, apiVersion string) string {
    if url[len(url)-1] != "/"[0] {
        url += "/"
    }
    if data[0] != "&"[0] {
        data = "&" + data
    }
    request := url + "api/" + apiVersion + "/" + url
    request += "?private_token=" + token
    request += data
    return sendGet(request)
}

func getSetting(setting string) (out string, err error) {
    cmd := exec.Command("git", "config", "--get", setting)
    raw_out, err := cmd.Output()
    out = strings.Trim(string(raw_out), "\n")
    return
}

func softGetSetting(setting string, desc string, badOptions [][2]string) (out string, newBadOptions [][2]string) {
    out, err := getSetting(setting)
    if err != nil {
        newBadOptions = append(badOptions, [2]string{setting, desc})
    } else {
        newBadOptions = badOptions
    }
    return
}

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

func varsFromGitConfig() (username string, url string, apiVersion string, token string, badOptions [][2]string) {
    username, badOptions = softGetSetting("gitlab.username", "gitlabusername", badOptions)
    url, badOptions = softGetSetting("gitlab.url", "http://my.gitlab.instance/", badOptions)
    if url[len(url)-1] != "/"[0] {
        url = url + "/"
    }
    apiVersion, badOptions = softGetSetting("gitlab.api", "v3", badOptions)
    token, badOptions = softGetSetting("gitlab.token", "your_gitlab_token", badOptions)
    return
}

func main() { //testing
    args, err := docopt.Parse(helpstring, argsToParse, automaticHelp, version, optionsFirst)
    if err != nil {
        panic(err)
    }
    username, url, apiVersion, token, badOptions := varsFromGitConfig()
    fmt.Println(args)
    if len(badOptions) != 0 {
        complainUndefined(badOptions)
        return
    }
    if username_opt, ok := args["-u"].(string); ok {
        username = username_opt
        badOptions = removeGood("gitlab.username", badOptions)
    }
    if url_opt, ok := args["-l"].(string); ok {
        url = url_opt
        badOptions = removeGood("gitlab.url", badOptions)
    }
    if api_opt, ok := args["-v"].(string); ok {
        apiVersion = api_opt
        badOptions = removeGood("gitlab.api", badOptions)
    }
    if token_opt, ok := args["-t"].(string); ok {
        token = token_opt
        badOptions = removeGood("gitlab.token", badOptions)
    }
    fmt.Println("username:", username)
    fmt.Println("api version:", apiVersion)
    fmt.Println("token:", token)
    fmt.Println("url:", url)
}
