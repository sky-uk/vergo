# Changelog

## [0.30.0] - 20-12-2024
Add Axion current behaviour when getting and bumping tag behind
Add NearestRelease flag to enable this new feature
Bumped goreleaser in order to release arm64 binaries

## [0.29.0] - 13-05-2024
Allows the environment variable key that is looked up to enable token based authentication to be configurable. The default is now `GH_TOKEN`.

## [0.28.1] - 09-05-2024
Fixed a bug with GITHUB_TOKEN authentication where pushes would fail when configured to use a GitHub token.
Fixed version of golangci-lint to work with more recent golang versions.


## [0.28.0] - 09-04-2024
Prefers to use the GITHUB_TOKEN Bearer authentication over SSH if the environment variable is present.
This is useful when pushing tags using GitHub actions.

## [0.27.0] - 08-02-2023
Fixed bug introduced in version `0.21.0` where `vergo bump` and `vergo check` would fail if the current commit is not 
the latest on a versioned branch e.g. `master` or `main`.

## [0.26.0] - 15-11-2022
Fixed bug in tag prefix trimming.

## [0.24.0] - 29-09-2022
Disable strict host checking using the global flag `--disable-strict-host-check` or `-d`. 
This is only intended to be used on CI where known_hosts is not cached. 

## [0.23.0] - 14-03-2022
`vergo get cv` should return `0.0.0-SNAPSHOT` in an empty repo or a repo without any tags 

## [0.22.0] - 10-03-2022
add ability extract release directives from the last commit message
e.g. : `vergo bump auto -t app` will look for patch/minor/major in commit message
if the latest commit message includes `[vergo:app:major-release]` string then auto will be translated to `major`

## [0.21.0] - 05-02-2022

`bump` should detect headless checkouts pointing to branches, `check` also should report the same issue

```
if vergo check release -t service; then
	version=$(vergo bump minor -t service)
else
	#bump would have failed because of some validation
	#this could be expected for branch builds, in this case push image to test with commit hash as image tag; don't bump/push any git tags
	version=$(git rev-parse --short HEAD)
fi
```
## [0.20.0] - 02-11-2021

add capability to check if a release can be skipped

## [0.19.0] - 02-11-2021

bump dependency versions

## [0.18.0] - 21-10-2021

Rename go.mod module name and required imports to follow remote go module path conventions 

## [0.17.0] - 08-10-2021

fail gracefully when no private keys available in the authentication agent

## [0.16.0] - 09-09-2021

Recognise tags with a slash prefix in order to support go multi-module projects

## [0.15.0] - 02-07-2021

Bump and Current version account should take account of both lightweight and annotated tags 

## [0.14.0] - 01-07-2021

current version should return tag on the HEAD if present

## [0.13.0] - 18-02-2021

automatically search for a local repository

## [0.12.0] - 18-02-2021

vergo umc-shared integration
