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