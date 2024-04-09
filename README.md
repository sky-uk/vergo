# Vergo

## Description

Vergo is a tool which allows you to easily attach semantic version tags to projects in Git multi-project repositories.
It's an alternative to Axion release Gradle plugin (https://github.com/allegro/axion-release-plugin.git) for non-Gradle projects.

## Comparison of axion-release-plugin and vergo

### axion-release-plugin

```
> git tag
project-0.1.0

> ./gradlew currentVersion
0.1.0

> git commit -m "Some commit."

> ./gradlew currentVersion
0.1.1-SNAPSHOT

> ./gradlew release

> git tag
project-0.1.0 project-0.1.1

> ./gradlew currentVersion
0.1.1

> ./gradlew publish
published project-0.1.1 release version

> ./gradlew markNextVersion -Prelease.version=1.0.0

> ./gradlew currentVersion
1.0.0-SNAPSHOT

```

### vergo

```
> git tag
project-0.1.0

> vergo get current-version -t project
0.1.0

> vergo get current-version -t project -p
project-0.1.0

> git commit -m "Some commit."

> vergo get current-version -t project
0.2.0-SNAPSHOT

> vergo bump patch -t project --log-level=error
0.1.1

> git tag
project-0.1.0
project-0.1.1

> vergo get current-version -t project
0.1.1

> vergo push -t project

>vergo bump patch -t project --push-tag
INFO[0000] Set tag project-0.1.2
0.1.2

```

## Installation

Prebuilt binaries for each release are available for the following platforms:

- Linux AMD64 and ARM64
- MacOS AMD64 and ARM64

You can find the binaries for the latest release at [https://github.com/sky-uk/vergo/releases/latest](https://github.com/sky-uk/vergo/releases/latest).

For other platforms, you can install the latest release of Go via `go install`:

```shell
go install github.com/sky-uk/vergo@latest
```

## Simple usages

* returns the latest tag/release prefixed with banana

  `vergo get latest-release --tag-prefix=banana`

* returns the previous tag/release prefixed with banana

  `vergo get previous-release --tag-prefix=banana`

* returns the current tag/release prefixed with banana, maybe a SNAPSHOT

  `vergo get current-version --tag-prefix=banana`

* increments patch part of the version prefixed with banana

  `vergo bump patch --tag-prefix=banana`

* increments minor part of the version prefixed with banana

  `vergo bump minor --tag-prefix=banana`

* increments minor part of the version prefixed with banana and pushes it to origin remote

  `vergo bump major --tag-prefix=apple --push-tag`

* pushes the tag to the remote as separate command

  `vergo push --tag-prefix=banana`

* supports the creation of tags with / seperated postfix will bump tags with the structure `orange/<major>.<minor>.<patch>`

  `vergo bump major --tag-prefix=orange/`

* checks if a release can be skipped by inspecting the latest commit message. If the commit message includes the hint `vergo:banana:skip-release` then the command fails saying release not required. 

  ```
  # expected usage 
  if vergo check release --tag-prefix=banana; then
    vergo bump major --tag-prefix=banana
  fi
  ```
* automatic increment by reading the last commit message. if the latest commit message includes `[vergo:app:major-release]` string then auto will be translated to `major`
  ```
    vergo bump auto -t app #will look for patch/minor/major in commit message
  ```

## Strict Host Checking

You can address the error `ssh: handshake failed: knownhosts: key is unknown ` when pushing tags with vergo in two ways:
- Calling `ssh-keyscan -H github.com >> ~/.ssh/known_hosts` prior to pushing your vergo tag to introduce github to your known hosts.
- Calling `vergo` with the `--disable-strict-host-check` flag. This should only be used on CI where known hosts are not cached.

## Using GITHUB_TOKEN inside GitHub Actions

Vergo will first try to use Token Bearer Authentication using the GITHUB_TOKEN environment variable when running inside a GitHub Action/Workflow. It will fallback to ssh based authentication if the GITHUB_TOKEN is not present.

Inside github actions please ensure that the GITHUB_TOKEN environment variable is set with the `${{ secrets.GITHUB_TOKEN }}` in order to push to the current repository.

Please see  [token authentication](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#using-the-github_token-in-a-workflow) for further details.

The GITHUB_TOKEN will require the following permissions to be able to push:
```yaml
    permissions:
      contents: write
```

## Running Locally - SSH Key Failures
You can address the error `FATA[0000] failed to get signers, make sure to add private key identities to the authentication agent  error="<nil>"` when pushing tags with vergo by:
- Calling `ssh-add ~/.ssh/<github_key>` to add your github ssh key to the ssh agent

## Contributions

We welcome contributions to Vergo. Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for details on how to contribute.
