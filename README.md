# Vergo

## Description

Vergo is an executable command line tool that is an alternative to axion gradle plugin
(https://github.com/allegro/axion-release-plugin.git)

# Simple usages

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


# Comparison of axion-release-plugin and vergo

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

* There is no replacement for `markNextVersion` at the moment. happy to implement this if it is required by other teams.
  Nova don't use it.