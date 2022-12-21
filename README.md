# testtrack-cli

[![Build status](https://github.com/Betterment/testtrack-cli/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Betterment/testtrack-cli/actions/workflows/ci.yml?query=branch%3Amain)

## TestTrack Split Config Management

Provides the `testtrack` command line interface as part of Betterment's [TestTrack](https://github.com/Betterment/test_track) open source split testing and feature gating platform.

The `testtrack` CLI has the following features:

* Provides an app-language-agnostic developer experience for configuring TestTrack splits from your shell alongside your app code.
* Manages TestTrack migrations representing those configuration changes within your application's codebase so that your split configs follow your code from development, through tests and into production via plugging `testtrack migrate` into your build/deploy pipeline.
* Provides a zero-config dependency-free simplified implementation of the TestTrack server REST API for developers to use when developing locally (or on a cloud VM/container if that's your cup of tea).

## Getting started

#### 1. Download/install testtrack CLI

Currently TestTrack binaries for linux and macOS are distributed via [GitHub releases](https://github.com/Betterment/testtrack-cli/releases). We'll add a homebrew tap soon to assist in installing and daemonizing the server on macOS.

#### 2. Initialize your project

From your app root directory run:

```bash
testtrack init_project
```

This will create a `testtrack` subdirectory in your app that will store migration and schema YAML files. *Commit these files to your repository.*

It will also create a `~/.testtrack` directory that will hold your local server's assignment overrides (`assignments.yml`).

#### 3. Set up your app name in .env

```bash
# [my_app_root]/.env
TESTTRACK_APP_NAME=my_app
```

This is the name that your app will use to authenticate with TestTrack server in production (and any other environments you support). Typically in a unirepo, this can be your app repo name.

#### 4. Point your app's TestTrack client at the testtrack fake server for development

You'll be embedding the same app name as you used above into the URL, e.g. `http://my_app@localhost:8297`

#### 5. Configure your application bootstrap scripts

You'll want to install the platform-appropriate `testtrack` binary and call `testtrack schema link --force` on each developer's machine.
We recommend [Scripts To Rule Them All](https://github.com/github/scripts-to-rule-them-all) as a pattern for bootstrapping app development environments for your team.

#### 6. Wire up testtrack to your build/deploy pipeline

a. Set an ENV var named `TESTTRACK_CLI_URL` with your TestTrack app credentials, e.g.:

```bash
TESTTRACK_CLI_URL=https://my_app:my_super_secret_app_token@testtrack.mydomain.com
```

into your build/deploy pipeline via whatever secrets management solution you use (heroku secrets, [sops](https://github.com/mozilla/sops), etc).

b. Make sure the platform-appropriate `testtrack` binary is installed in your build/deploy environment (e.g. Jenkins, GoCD)

c. Run `testtrack migrate` from your app root. For server-side apps, this is great to wire up to the same build pipeline phase where you'd apply database migrations for your app. For mobile or other client-side apps, you'll want to run it after tests have passed and before persisting your gold master build artifact.

#### 7. Start creating splits!

By default, splits will default to 100% false:

```bash
testtrack create feature_gate my_new_feature_q2_2019_enabled
```

Use the `--weights` argument to specify custom variants and weights:

```
testtrack create experiment my_new_feature_q2_2019_experiment --weights "control: 50, treatment_a: 25, treatment_b: 25"
```

#### 8. Retire splits

Once an experiment is finished or feature released, remove all references to split in code. Then, decide and retire split.

```bash
testtrack destroy split my_new_feature_q2_2019_enabled --decision=true
```

Run `testtrack help` for more documentation on how to configure splits and other TestTrack resources.

Happy TestTracking!

## How to Contribute

We would love for you to contribute! Anything that benefits the majority of TestTrack users—from a documentation fix to an entirely new feature—is encouraged.

Before diving in, [check our issue tracker](https://github.com/Betterment/testtrack-cli/issues) and consider creating a new issue to get early feedback on your proposed change.

### Suggested Workflow

* Fork the project and create a new branch for your contribution.
* Write your contribution (and any applicable test coverage).
* Make sure all tests pass (`make test`).
* Submit a pull request.

### Some tips for those new to golang

* Set up your workspace according to [go standards](https://golang.org/doc/code.html#Organization).
* For macOS and homebrew users, run `brew bundle` to install `go` itself.
* Build and run the CLI using `go run testtrack/main.go`.
