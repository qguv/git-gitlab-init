/*
Initializes a repository remotely for GitLab-hosted origin servers.
*/

package main

import (
    "fmt"
    "github.com/docopt/docopt.go"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "strings"
)

// Variables about gitlab-init
var (
    version       string   = "0.1.1"
    docopt_argument_source   []string = os.Args[1:]
    docopt_autohelp_enabled bool     = true
    docopt_optionsfirst_enabled  bool     = true
    docopt_usage_pattern    string   = `git gitlab-init

Create an empty Git repository on GitLab and locally.

The commands below can be used as "git gitlab-init" or as "git-gitlab-init".

Usage:
  git-gitlab-init (-h | --help | --version)
  git-gitlab-init [-p PRIVACYLEVEL] [-u USERNAME] [-l URL] [-d DESCRIPTION]
                  [-v API_VERSION] [-t API_TOKEN] [--debug] [--] <repository>

Arguments:
  <repository>      Specify repository name.

Options:
  -h, --help        Shows this screen and exit.
  -p PRIVACYLEVEL   Sets viewing permission status of repository. Valid options
                    are public, private, or internal. [default: private]
  -d DESCRIPTION    Specify repository description.
  -u USERNAME       Specify Gitlab username.
  -l URL            Specify Gitlab instance url.
  -v API_VERSION    Specify Gitlab api version url. [default: v3]
  -t API_TOKEN      Specify Gitlab api token, found in your Gitlab profile settings.
  --debug           You hopefully won't need this option.
  --version         Prints version and exits.
`
)

// findInSlice searches the given slice for a 2-array where element 0 equals the
// given string and returns its index.
func findInSlice(s string, slice [][2]string) int {
    for i, v := range slice {
        if v[0] == s {
            return i
        }
    }
    return -1
}

// removeFromSlice removes the element with the given index from the given
// slice and returns the resulting slice.
func removeFromSlice(i int, slice [][2]string) [][2]string {
    return append(slice[:i], slice[i+1:]...)
}

// removeGood removes the first 2-array it finds (in the given slice) where
// element 0 equals the given string.
func removeGood(s string, slice [][2]string) [][2]string {
    i := findInSlice(s, slice)
    if i != -1 {
        return removeFromSlice(i, slice)
    }
    return [][2]string{}
}

// scrubUrl adds a trailing backslash to the given url if there isn't one
// already.
func scrubUrl(webaddress string) string {
    if webaddress[len(webaddress)-1] != "/"[0] {
        webaddress += "/"
    }
    return webaddress
}

// runCommand executes a shell command and returns the resulting stdout as a
// string and any errors as errors.
func runCommand(name string, arg ...string) (string, error) {
    cmd := exec.Command(name, arg...)
    raw_out, err := cmd.Output()
    out := string(raw_out)
    return out, err
}

// initialize initializes a git repository locally with a barebones README.md
// and a basic first commit.
func initialize(project_name string, origin_username string, origin_root_address string) (ok bool) {
    runCommand("git", "init")

    f, err := os.Create("README.md")
    defer f.Close()
    if err != nil {
        panic(err)
    }
    _, err = f.WriteString("# " + project_name)
    if err != nil {
        panic(err)
    }

    origin_root_address = scrubUrl(origin_root_address)
    origin_relative_address := origin_username + "/" + project_name + ".git"
    origin_full_address := origin_root_address + origin_relative_address

    runCommand("git", "add", "README.md")
    runCommand("git", "commit", "-m", "initial commit")
    runCommand("git", "remote", "add", "origin", origin_full_address)
    runCommand("git", "push", "-u", "origin", "master")

    // #TODO: detect problems with runCommand and throw errors

    return true // ok
}

// sendPost sends a HTTP POST request to the given url with given payload.
func sendPost(address string, payload url.Values) string {
    resp, err := http.PostForm(address, payload)
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

// apiCommand sends a HTTP POST request containing the given payload to the
// Gitlab instance with the given API credentials.
func apiCommand(gitlab_root_address string, api_subdirectory string, payload url.Values, api_token string, gitlab_api_version string) string {
    gitlab_root_address = scrubUrl(gitlab_root_address)
    api_request_address := gitlab_root_address + "api/" + gitlab_api_version + "/" + api_subdirectory
    api_request_address += "?private_token=" + api_token
    return sendPost(api_request_address, payload)
}

// makeRemoteRepo initializes a repo to the Gitlab instance with the given API
// credentials.
func makeRemoteRepo(gitlab_root_address string, repo_name string, repo_description string, api_token string, gitlab_api_version string, repo_permissions string) string {
    payload := make(url.Values)
    payload.Set("name", repo_name)
    if repo_description != "" {
        payload.Set("description", repo_description)
    }
    if repo_permissions == "private" {
        payload.Set("visibility_level", "0")
    } else if repo_permissions == "public" {
        payload.Set("visibility_level", "20")
    } else if repo_permissions == "internal" {
        payload.Set("visibility_level", "10")
    } else {
        return "ERROR! in repo_permissions" // #DEBUG
    }
    return apiCommand(gitlab_root_address, "projects", payload, api_token, gitlab_api_version)
}

// getSetting scrapes `git config` for the given setting and outputs its value.
func getSetting(setting string) (value string, err error) {
    setting_cmd := exec.Command("git", "config", "--get", setting)
    raw_value, err := setting_cmd.Output()
    value = strings.Trim(string(raw_value), "\n")
    return
}

// softGetSetting scrapes `git config` for the given setting and outputs its
// value if it exists; otherwise the option and a short description are added
// to a running slice of undefined options.
func softGetSetting(setting string, setting_help_description string, bad_options [][2]string) (value string, new_bad_options [][2]string) {
    value, err := getSetting(setting)
    if err != nil {
        new_bad_options = append(bad_options, [2]string{setting, setting_help_description})
    } else {
        new_bad_options = bad_options
    }
    return
}

// complainUndefined prints a pretty error message explaining how to remedy
// configuration issues à la the official git subcommands.
func complainUndefined(bad_options [][2]string) {
    fmt.Println("Error! Your Gitlab API settings aren't defined.")
    fmt.Println("Try running the following:\n")
    var param [2]string
    for _, param = range bad_options {
        setting, desc := param[0], param[1]
        fmt.Println("    git config --global " + setting + " \"" + desc + "\"")
    }
    fmt.Println("")
}

// varsFromGitConfig scrapes `git config` for all the necessary gitlab API
// authentication information
func varsFromGitConfig() (gitlab_username string, gitlab_root_address string, gitlab_api_version string, gitlab_api_token string, bad_options [][2]string) {
    gitlab_username, bad_options = softGetSetting("gitlab.username", "gitlabusername", bad_options)
    gitlab_root_address, bad_options = softGetSetting("gitlab.url", "http://my.gitlab.instance/", bad_options)
    gitlab_root_address = scrubUrl(gitlab_root_address)
    gitlab_api_version, bad_options = softGetSetting("gitlab.api", "v3", bad_options)
    gitlab_api_token, bad_options = softGetSetting("gitlab.token", "your_gitlab_token", bad_options)
    return
}

func main() {
    args, err := docopt.Parse(docopt_usage_pattern, docopt_argument_source, docopt_autohelp_enabled, version, docopt_optionsfirst_enabled)
    if err != nil {
        panic(err)
    }

    // reflect debug flag as bool and store in debug
    debug := args["--debug"].(bool)

    // Scraping config files for API credentials
    gitlab_username, gitlab_root_address, gitlab_api_version, gitlab_api_token, bad_options := varsFromGitConfig()
    if debug {
        fmt.Println(args)
    }

    // Overriding config parameters from docopt
    if gitlab_username_opt, ok := args["-u"].(string); ok {
        gitlab_username = gitlab_username_opt
        bad_options = removeGood("gitlab.username", bad_options)
    }
    if gitlab_root_address_opt, ok := args["-l"].(string); ok {
        gitlab_root_address = gitlab_root_address_opt
        bad_options = removeGood("gitlab.url", bad_options)
    }
    if api_opt, ok := args["-v"].(string); ok {
        gitlab_api_version = api_opt
        bad_options = removeGood("gitlab.api", bad_options)
    }
    if gitlab_api_token_opt, ok := args["-t"].(string); ok {
        gitlab_api_token = gitlab_api_token_opt
        bad_options = removeGood("gitlab.token", bad_options)
    }

    gitlab_root_address = scrubUrl(gitlab_root_address)

    if len(bad_options) != 0 {
        complainUndefined(bad_options)
        return
    }

    // Setting repo settings from docopt
    var repo_description string
    if repo_description_opt, ok := args["-d"].(string); ok {
        repo_description = repo_description_opt
    } else {
        repo_description = ""
    }
    repo_name := args["<repository>"].(string)
    repo_permissions := args["-p"].(string)

    if debug {
        fmt.Println("username:", gitlab_username)
        fmt.Println("api version:", gitlab_api_version)
        fmt.Println("token:", gitlab_api_token)
        fmt.Println("url:", gitlab_root_address)
    }

    response := makeRemoteRepo(gitlab_root_address, repo_name, repo_description, gitlab_api_token, gitlab_api_version, repo_permissions)
    if debug {
        fmt.Println(response)
    }

    initialize(repo_name, gitlab_username, gitlab_root_address)
}
