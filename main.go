// This is a trivial application which will output a dump of repositories
// which are hosted upon github, or some other host which uses a
// compatible API.
//
// It relies upon having an access-token for authentication.

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)

var (
	//
	// Context for all calls
	//
	ctx context.Context

	//
	// The actual github client
	//
	client *github.Client

	//
	// The token to use for accessing the remote host.
	//
	// This is required because gitbucket prefers to see
	//
	//     Authorization: token SECRET-TOKEN
	//
	// Instead of:
	//
	//     Authorization: bearer SECRET-TOKEN
	//
	oauthToken = &oauth2.Token{}

	//
	// The number of repos to fetch from the API at a time.
	//
	pageSize = 50

	//
	// Our version number, set for release-builds.
	//
	version = "unreleased"
)

// Login accepts the address of a github endpoint, and a corresponding
// token to authenticate with.
//
// We use the login to get the user-information which confirms
// that the login was correct.
func Login(api string, token string) error {

	// Setup context
	ctx = context.Background()

	// Setup token
	ts := oauth2.StaticTokenSource(oauthToken)
	tc := oauth2.NewClient(ctx, ts)

	// Create the API-client
	client = github.NewClient(tc)

	// If the user is using a custom URL which doesn't have the
	// versioned API-suffix add it.  This appears to be necessary.
	if api != "https://api.github.com/" {
		if !strings.HasSuffix(api, "/api/v3/") {
			if !strings.HasSuffix(api, "/") {
				api += "/"
			}
			api += "api/v3/"
		}
	}

	// Parse the URL for sanity, and update the client with it
	url, err := url.Parse(api)
	if err != nil {
		return err
	}
	client.BaseURL = url

	// Fetch user-information about the user who we are logging in as.
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return err
	}

	// Ensure we have a login
	if *user.Login == "" {
		return fmt.Errorf("we failed to find our username, which suggests our login failed")
	}

	return nil
}

// getPersonalRepos returns all the personal repositories which
// belong to our user.
func getPersonalRepos(fetch string) ([]*github.Repository, error) {

	var results []*github.Repository

	// Fetch in pages
	opt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: pageSize},
		Type:        fetch,
	}

	// Loop until we're done.
	for {
		repos, resp, err := client.Repositories.List(ctx, "", opt)
		if err != nil {
			return results, err
		}
		results = append(results, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return results, nil

}

// getOrganizationalRepositores finds all the organizations the
// user is a member of, then fetches their repositories
func getOrganizationalRepositores(fetch string) ([]*github.Repository, error) {

	var results []*github.Repository

	// Get the organizations the user is a member of.
	orgs, _, err := client.Organizations.List(ctx, "", nil)
	if err != nil {
		return results, err
	}

	// Fetch in pages
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: pageSize},
		Type:        fetch,
	}

	// For each organization we want to get their repositories.
	for _, org := range orgs {

		// Loop forever getting the repositories
		for {

			repos, resp, err := client.Repositories.ListByOrg(ctx, *org.Login, opt)
			if err != nil {
				return results, err
			}
			results = append(results, repos...)
			if resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}
	}

	return results, nil
}

//
// Entry-point
//
func main() {

	//
	// Parse flags
	//
	archived := flag.Bool("archived", false, "Include archived repositories in the output?")
	api := flag.String("api", "https://api.github.com/", "The API end-point to use for the remote git-host.")
	authHeader := flag.Bool("auth-header-token", false, "Use an authorization-header including 'token' rather than 'bearer'.\nThis is required for gitbucket, and perhaps other systems.")
	exclude := flag.String("exclude", "", "Comma-separated list of repositories to exclude.")
	getOrgs := flag.String("organizations", "all", "Which organizational repositories to fetch.\nValid values are 'public', 'private', 'none', or 'all'.")
	getPersonal := flag.String("personal", "all", "Which personal repositories to fetch.\nValid values are 'public', 'private', 'none', or 'all'.")
	http := flag.Bool("http", false, "Generate HTTP-based clones rather than SSH-based ones.")
	ssh := flag.Bool("ssh", false, "Add 'ssh://'-prefix to the git clone command.")
	output := flag.String("output", "", "Write output to the named file, instead of printing to STDOUT.")
	prefix := flag.String("prefix", "", "The prefix beneath which to store the repositories upon the current system.")
	token := flag.String("token", "", "The API token used to authenticate to the remote API-host.")
	versionCmd := flag.Bool("version", false, "Report upon our version, and terminate.")
	flag.Parse()

	//
	// Showing only the version?
	//
	if *versionCmd {
		fmt.Printf("github2mr %s\n", version)
		return
	}

	//
	// Validate the repository-types
	//
	if *getPersonal != "all" &&
		*getPersonal != "none" &&
		*getPersonal != "public" &&
		*getPersonal != "private" {
		fmt.Fprintf(os.Stderr, "Valid settings are 'public', 'private', 'none', or 'all'\n")
		return
	}
	if *getOrgs != "all" &&
		*getOrgs != "none" &&
		*getOrgs != "public" &&
		*getOrgs != "private" {
		fmt.Fprintf(os.Stderr, "Valid settings are 'public', 'private', 'none', or 'all'\n")
		return
	}

	//
	// Get the authentication token supplied via the flag, falling back
	// to the environment if nothing has been specified.
	//
	tok := *token
	if tok == "" {
		// Fallback
		tok = os.Getenv("GITHUB_TOKEN")

		if tok == "" {
			fmt.Printf("Please specify your github token!\n")
			return
		}
	}

	//
	// Populate our global OAUTH token with the supplied value.
	//
	oauthToken.AccessToken = tok

	//
	// Allow setting the authorization header-type, if required.
	//
	if *authHeader {
		oauthToken.TokenType = "token"
	}

	//
	// Login and confirm that this worked.
	//
	err := Login(*api, tok)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login error - is your token set/correct? %s\n", err.Error())
		return
	}

	//
	// Fetch details of all "personal" repositories, unless we're not
	// supposed to.
	//
	var personal []*github.Repository
	if *getPersonal != "none" {
		personal, err = getPersonalRepos(*getPersonal)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch personal repository list: %s\n", err.Error())
			return
		}
	}

	//
	// Fetch details of all organizational repositories, unless we're
	// not supposed to.
	//
	var orgs []*github.Repository
	if *getOrgs != "none" {
		orgs, err = getOrganizationalRepositores(*getOrgs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch organizational repositories: %s\n",
				err.Error())
			return
		}
	}

	//
	// If the prefix is not set then create a default.
	//
	// This will be of the form:
	//
	//    ~/Repos/github.com/x/y
	//    ~/Repos/git.example.com/x/y
	//    ~/Repos/git.steve.fi/x/y
	//
	// i.e "~/Repos/${git host}/${owner}/${path}
	//
	// (${git host} comes from the remote API host.)
	//
	repoPrefix := *prefix
	if repoPrefix == "" {

		// Get the hostname
		url, _ := url.Parse(*api)
		host := url.Hostname()

		// Handle the obvious case
		if host == "api.github.com" {
			host = "github.com"
		}

		// Generate a prefix
		repoPrefix = os.Getenv("HOME") + "/Repos/" + host
	}

	//
	// Combine the results of the repositories we've found.
	//
	var all []*github.Repository
	all = append(all, personal...)
	all = append(all, orgs...)

	//
	// Sort the list, based upon the full-name.
	//
	sort.Slice(all[:], func(i, j int) bool {

		// Case-insensitive sorting.
		a := strings.ToLower(*all[i].FullName)
		b := strings.ToLower(*all[j].FullName)

		return a < b
	})

	//
	// Repos we're excluding
	//
	excluded := strings.Split(*exclude, ",")

	//
	// Structure we use for template expansion
	//
	type Repo struct {
		// Prefix-directory for local clones.
		Prefix string

		// Name of the repository "owner/repo-name".
		Name string

		// Source to clone from http/ssh-based.
		Source string
	}

	//
	// Repos we will output
	//
	var repos []*Repo

	//
	// Now format the repositories we've discovered.
	//
	for _, repo := range all {

		//
		// If the repository is archived then
		// skip it, unless we're supposed to keep
		// it.
		//
		if *repo.Archived && !*archived {
			continue
		}

		//
		// The clone-type is configurable
		//
		clone := *repo.SSHURL
		if *http {
			clone = *repo.CloneURL
		}

		//
		// Sometimes SSH clones need a prefix
		//
		if *ssh {
			clone = "ssh://" + clone
		}

		//
		// Hack!
		//
		clone = strings.ReplaceAll(clone, ":4444:", ":4444/")

		//
		// Should we exclude this entry?
		//
		skip := false
		for _, exc := range excluded {

			exc = strings.TrimSpace(exc)

			if len(exc) > 0 && strings.Contains(strings.ToLower(clone), strings.ToLower(exc)) {
				skip = true
			}
		}

		// Skipped
		if skip {
			continue
		}

		repos = append(repos, &Repo{Prefix: repoPrefix,
			Name:   *repo.FullName,
			Source: clone})
	}

	//
	// Load the template we'll use for formatting the output
	//
	tmpl := `# Generated by github2mr - {{len .}} repositories

{{range .}}
[{{.Prefix}}/{{.Name}}]
checkout = git clone {{.Source}}
{{end}}
`

	//
	// Parse the template and execute it.
	//
	var out bytes.Buffer
	t := template.Must(template.New("tmpl").Parse(tmpl))
	err = t.Execute(&out, repos)

	//
	// If there were errors we're done.
	//
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error interpolating template:%s\n", err.Error())
		return
	}

	//
	// Show the results, or write to the specified file as appropriate
	//
	if *output != "" {
		file, err := os.Create(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open %s:%s\n", *output, err.Error())
			return
		}
		defer file.Close()
		file.Write(out.Bytes())
	} else {
		fmt.Println(out.String())
	}

	//
	// All done.
	//
}
