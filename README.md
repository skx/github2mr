# Configuration / Usage

Once installed you'll need to configure your github token, which you can generate from [withing your github settings](https://github.com/settings/tokens).

you can either pass the token as an argument to the tool (via `github2mr -token=xxxxx`), or store it in the environment in the variable GITHUB_TOKEN:

     $ export GITHUB_TOKEN=xxxxx
     $ github2mr [options]

You can run `github2mr -help` to see available options, but in brief:

* You can choose a default prefix to clone your repositories to.
  * By default all repositories will be located at `~/Repos/${git_host}`.
* You can exclude all-organizational repositories.
  * Or the reverse, ignoring all personal-repositories.
* You can exclude repositories by name.
* You can default to cloning repositories via HTTP, instead of SSH.
* By default all _archived_ repositories are excluded.


## Other Git Hosts

This tool can be configured to point at other systems which use the same
API as the public-facing Github site.

To use it against a self-hosted Github Enterprise installation, for example,
simply specify the URL:

     $ export GITHUB_TOKEN=xxxxx
     $ github2mr -api=https://git.example.com/ [options]

It has also been tested against an installation of [gitbucket](https://github.com/gitbucket/gitbucket) which can be configured a similar way - however in this case you'll find that you receive an error "401 bad credentials" unless you add the `-auth-header-token` flag:

      $ export GITHUB_TOKEN=xxxxx
      $ github2mr -api=https://git.example.com/ -auth-header-token

This seems to be related to the OAUTH header the library I'm using sends, by default it will send a HTTP request looking like this:

```
GET /api/v3/users/skx/repos HTTP/1.1
Host: localhost:9999
User-Agent: go-github
Accept: application/vnd.github.mercy-preview+json
Authorization: Bearer SECRET-TOKEN
Accept-Encoding: gzip
```

Notice that the value of the `Authorization`-header begins with `Bearer`?  Gitbucket prefers to see `Authorization: token SECRET-VALUE-HERE`.




# Github Setup

This repository is configured to run tests upon every commit, and when pull-requests are created/updated.  The testing is carried out via [.github/run-tests.sh](.github/run-tests.sh) which is used by the [github-action-tester](https://github.com/skx/github-action-tester) action.

Releases are automated in a similar fashion via [.github/build](.github/build), and the [github-action-publish-binaries](https://github.com/skx/github-action-publish-binaries) action.

Currently these are reporting failures; but I'm in the process of fixing them.



Steve
--
