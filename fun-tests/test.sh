#!/usr/bin/env bash

set -ex
export PS4='$0.$LINENO '

finish() {
  local rc=$?
  if ((rc == 0)); then
    echo "OK"
  else
    echo "***FAIL***"
  fi
  exit $rc
}
trap finish EXIT
trap 'exit 2' HUP INT QUIT TERM

#setup clones a test repo to be used for assertions
setUp() {
  cd /tmp
  rm -rf /tmp/vergo-test-repo /tmp/vergo-test-repo-clone || true
  git clone git@github.com:sky-uk/vergo-test-repo.git /tmp/vergo-test-repo
  git clone /tmp/vergo-test-repo /tmp/vergo-test-repo-clone
  cd /tmp/vergo-test-repo-clone
  [[ "$(git tag -l apple-0.2.0 app-0.1.1 banana-2.0.0)" == "" ]]
}

setUp
readonly VERGO="$PWD/bin/vergo"
readonly head="$(git rev-parse HEAD)"
readonly headShort="$(git rev-parse --short HEAD)"
readonly vergoVersion="$(VERGO version simple)"

#test get latest-release
[[ "$(VERGO get lr --tag-prefix=apple)" == "0.1.1" ]]
[[ "$(VERGO get lr --tag-prefix=app)" == "0.1.0" ]]
[[ "$(VERGO get lr --tag-prefix=banana)" == "1.1.2" ]]

[[ "$(VERGO get lr --tag-prefix=apple --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {0.1.1 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/apple-0.1.1}\n"
0.1.1' ]]
[[ "$(VERGO get lr --tag-prefix=app --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {0.1.0 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/app-0.1.0}\n"
0.1.0' ]]

#test get current-version
[[ "$(VERGO get cv --tag-prefix=apple)" == "0.2.0-SNAPSHOT" ]]
[[ "$(VERGO get cv --tag-prefix=app)" == "0.2.0-SNAPSHOT" ]]

#test get current-version with prefix included in the output
[[ "$(VERGO get cv --tag-prefix=apple -p)" == "apple-0.2.0-SNAPSHOT" ]]
[[ "$(VERGO get cv --tag-prefix=app -p)" == "app-0.2.0-SNAPSHOT" ]]

#test get current-version for SNAPSHOT
[[ "$(VERGO get cv --tag-prefix=apple --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {0.1.1 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/apple-0.1.1}\n"
0.2.0-SNAPSHOT' ]]
[[ "$(VERGO get cv --tag-prefix=app --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {0.1.0 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/app-0.1.0}\n"
0.2.0-SNAPSHOT' ]]

#test bump minor first tag, melon prefix dees not exist in repo
[[ "$(VERGO bump minor -tmelon --dry-run --log-level=error 2>&1)" == "0.1.0" ]]

#test bump minor
[[ "$(VERGO bump minor --tag-prefix=apple --log-level=trace 2>&1)" == *'Set tag apple-0.2.0'* ]]
[[ "$(VERGO bump minor --tag-prefix=apple --log-level=trace 2>&1)" == *'Push not enabled'* ]]
[[ "$(VERGO bump minor --tag-prefix=apple --log-level=error 2>&1)" == '0.2.0' ]]
#test bump minor check version
[[ "$(VERGO get lr --tag-prefix=apple)" == "0.2.0" ]]
[[ "$(VERGO get cv -tapple)" == "0.2.0" ]]

#test bump patch
[[ "$(VERGO bump patch --tag-prefix=app --log-level=trace 2>&1)" == *'Set tag app-0.1.1'* ]]
[[ "$(VERGO bump patch --tag-prefix=app --log-level=trace 2>&1)" == *'Push not enabled'* ]]
[[ "$(VERGO bump patch --tag-prefix=app --log-level=error 2>&1)" == "0.1.1" ]]
#test bump patch check version
[[ "$(VERGO get lr -t=app)" == "0.1.1" ]]
[[ "$(VERGO get cv --tag-prefix=app)" == "0.1.1" ]]

#test bump major
readonly tagPrefixBananaVersion_2_0_0="$(VERGO bump major --push-tag --tag-prefix=banana --log-level=trace 2>&1)"
[[ $tagPrefixBananaVersion_2_0_0 == *'Set tag banana-2.0.0'* ]]
[[ $tagPrefixBananaVersion_2_0_0 == *'Pushing tag: banana-2.0.0'* ]]
#test bump major check version
[[ "$(VERGO get lr --tag-prefix=banana)" == "2.0.0" ]]
[[ "$(VERGO get cv --tag-prefix=banana)" == "2.0.0" ]]

#test list tags with tag-prefix=app
[[ "$(VERGO list --tag-prefix=app --log-level=trace 2>&1)" == *'0.1.1
0.1.0' ]]
[[ "$(VERGO list --tag-prefix=app --log-level=trace --sort-direction asc 2>&1)" == *'0.1.0
0.1.1' ]]
#test list tags with tag-prefix=apple
[[ "$(VERGO list --tag-prefix=apple --log-level=trace 2>&1)" == *'0.2.0
0.1.1' ]]
[[ "$(VERGO list --tag-prefix=apple --log-level=trace --sort-direction asc 2>&1)" == *'0.1.1
0.2.0' ]]

#test list tags with tag-prefix=app with prefix included
[[ "$(VERGO list --tag-prefix=app -p --log-level=trace 2>&1)" == *'app-0.1.1
app-0.1.0' ]]
[[ "$(VERGO list --tag-prefix=app -p --log-level=trace --sort-direction asc 2>&1)" == *'app-0.1.0
app-0.1.1' ]]

#test list tags with tag-prefix=apple with prefix included
[[ "$(VERGO list --tag-prefix=apple -p --log-level=trace 2>&1)" == *'apple-0.2.0
apple-0.1.1' ]]
[[ "$(VERGO list --tag-prefix=apple -p --log-level=trace --sort-direction asc 2>&1)" == *'apple-0.1.1
apple-0.2.0' ]]

#test push
[[ "$(VERGO push --tag-prefix=apple --log-level=trace 2>&1)" == *'Pushing tag: apple-0.2.0'* ]]
[[ "$(VERGO push --tag-prefix=apple --log-level=trace 2>&1)" == *'Pushing tag: apple-0.2.0'* ]]
[[ "$(VERGO push --tag-prefix=apple --log-level=trace 2>&1)" == *'origin remote was up to date, no push done'* ]]
[[ "$(VERGO push --tag-prefix=app --log-level=trace 2>&1)" == *'Pushing tag: app-0.1.1'* ]]

#test push tags are present on remote
remote_tags=$(git ls-remote --tags origin apple-0.2.0 app-0.1.1 banana-2.0.0)
[[ "${remote_tags}" == *'refs/tags/app-0.1.1'* ]]
[[ "${remote_tags}" == *'refs/tags/apple-0.2.0'* ]]
[[ "${remote_tags}" == *'refs/tags/banana-2.0.0'* ]]

#test headless checkout
setUp
git checkout 117443bb0d121fa75bbde8b4c75bfadebf90c954 2>/dev/null 1>&2
[[ "$(VERGO bump minor --tag-prefix=apple --log-level=trace --versioned-branch-names main 2>&1)" == *'command disabled for branches'* ]]
[[ "$(VERGO bump minor --tag-prefix=apple --log-level=trace --versioned-branch-names HEAD 2>&1)" == *'Set tag apple-0.2.0'* ]]
