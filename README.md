
# Contents

* [github2mr](#github2mr)
  * [Brief mr Example](#brief-mr-example)
* [Installation](#installation)
* [Configuration / Usage](#configuration--usage)
  * [Other Git Hosts](#other-git-hosts)
* [Github Setup](#github-setup)




# github2mr

Many [Github](https://github.com/) users have a large number of repositories upon which they work, and managing them all can sometimes be difficult.

One excellent tool which helps a lot is the [myrepos](https://myrepos.branchable.com/) package, containing a binary named `mr`, which allows you to run many operations upon multiple repositories with one command.  (It understands git, mercurial, darcs, cvs, and many other types of revision-control systems.)

This repository contains a simple command-line client which allows you to easily generate a configuration file containing __all__ your github repositories fetching them via the [Github API](https://developer.github.com/v3/) with various filtering and limiting options.

The end result of using `mr` and `github2mr` is that you should be able to clone all your remote github repositories, and update them easily with only a couple of commands which is great for when you work/develop/live on multiple machines.


## Brief `mr` Example

Let us pretend I'm moving to a new machine; first of all I export the list of all my remote repositories to a configuration file using _this_ tool:

    github2mr > ~/Repos/.mrconfig.github

* **NOTE**: The first time you create a new configuration file you will need to mark it as being trusted, because it is possible for configuration files to contain arbitrary shell-commands.
  * Mark the configuration file as trusted by adding it's name to `~/.mrtrust`:
      * `echo ~/Repos/.mrconfig.github >> ~/.mrtrust`

Now that we've populated a configuration-file we can tell `mr` to checkout each of those repositories:

    mr --jobs 8 --config ~/Repos/.mrconfig.github

Later in the week I can update all the repositories which have been cloned, pulling in any remote changes that have been made from other systems:

    mr --jobs 8 --config ~/Repos/.mrconfig.github update

**NOTE**: If you prefer you can just use `update` all the time, `mr` will checkout a repository if it is missing as part of the `update` process.  I'm using distinct flags here for clarity.  Please read the `mr`-manpage to look at the commands it understands.


# Installation

You should be able to install this application using the standard golang approach:

    $ go get github.com/skx/github2mr

If you prefer you can [download the latest binary](http://github.com/skx/github2mr/releases) release, for various systems.




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
