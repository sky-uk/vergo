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
  [[ "$(git tag -l apple-0.2.0 app-0.1.1 banana-2.0.0 orange/v3.0.0)" == "" ]]
}

setUpLikeCI() {
  cd /tmp
  rm -rf /tmp/vergo-test-repo /tmp/vergo-test-repo-clone || true
  git init /tmp/vergo-test-repo
  git -C /tmp/vergo-test-repo fetch --tags -- git@github.com:sky-uk/vergo-test-repo.git +refs/heads/*:refs/remotes/origin/*
  git -C /tmp/vergo-test-repo checkout "$(git -C /tmp/vergo-test-repo show-ref --hash refs/remotes/origin/master)"
  git -C /tmp/vergo-test-repo config remote.origin.url git@github.com:sky-uk/vergo-test-repo.git

  cd /tmp/vergo-test-repo
  [[ "$(git tag -l apple-0.2.0 app-0.1.1 banana-2.0.0 orange/v3.0.0)" == "" ]]
}

readonly head="$(git rev-parse HEAD)"
readonly headShort="$(git rev-parse --short HEAD)"
readonly vergoVersion="$(vergo version simple)"
setUp

#test check release
vergo check release --tag-prefix=apple
vergo check release --tag-prefix=app
vergo check release --tag-prefix=banana
if vergo check release --tag-prefix=cherry; then
  false
else
  true
fi
[[ "$(vergo check release --tag-prefix=cherry 2>&1)" == 'Error: skip release hint present: cherry' ]]

#test get latest-release
[[ "$(vergo get lr --tag-prefix=apple)" == "0.1.1" ]]
[[ "$(vergo get lr --tag-prefix=app)" == "0.1.0" ]]
[[ "$(vergo get lr --tag-prefix=banana)" == "1.1.2" ]]
[[ "$(vergo get lr --tag-prefix=orange/v)" == "2.3.0" ]]

[[ "$(vergo get lr --tag-prefix=apple --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {0.1.1 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/apple-0.1.1}\n"
0.1.1' ]]
[[ "$(vergo get lr --tag-prefix=app --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {0.1.0 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/app-0.1.0}\n"
0.1.0' ]]
[[ "$(vergo get lr --tag-prefix=orange/v --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {2.3.0 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/orange/v2.3.0}\n"
2.3.0' ]]

#test get current-version
[[ "$(vergo get cv --tag-prefix=apple)" == "0.2.0-SNAPSHOT" ]]
[[ "$(vergo get cv --tag-prefix=app)" == "0.2.0-SNAPSHOT" ]]
[[ "$(vergo get cv --tag-prefix=orange/v)" == "2.4.0-SNAPSHOT" ]]

#test get current-version with prefix included in the output
[[ "$(vergo get cv --tag-prefix=apple -p)" == "apple-0.2.0-SNAPSHOT" ]]
[[ "$(vergo get cv --tag-prefix=app -p)" == "app-0.2.0-SNAPSHOT" ]]
[[ "$(vergo get cv --tag-prefix=orange/v -p)" == "orange/v2.4.0-SNAPSHOT" ]]

#test get current-version for SNAPSHOT
[[ "$(vergo get cv --tag-prefix=apple --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {0.1.1 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/apple-0.1.1}\n"
0.2.0-SNAPSHOT' ]]
[[ "$(vergo get cv --tag-prefix=app --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {0.1.0 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/app-0.1.0}\n"
0.2.0-SNAPSHOT' ]]
[[ "$(vergo get cv --tag-prefix=orange/v --log-level=trace 2>&1)" == *'level=debug msg="Latest version: {2.3.0 b26954d9cb20e58b5d6d9b9c1930ae998e8b6e3c refs/tags/orange/v2.3.0}\n"
2.4.0-SNAPSHOT' ]]

#test bump minor first tag, melon prefix dees not exist in repo
[[ "$(vergo bump minor -tmelon --dry-run --log-level=error 2>&1)" == "0.1.0" ]]

#test bump minor
[[ "$(vergo bump minor --tag-prefix=apple --log-level=trace 2>&1)" == *'Set tag apple-0.2.0'* ]]
[[ "$(vergo bump minor --tag-prefix=apple --log-level=trace 2>&1)" == *'Push not enabled'* ]]
[[ "$(vergo bump minor --tag-prefix=apple --log-level=error 2>&1)" == '0.2.0' ]]
#test bump minor check version
[[ "$(vergo get lr --tag-prefix=apple)" == "0.2.0" ]]
[[ "$(vergo get cv -tapple)" == "0.2.0" ]]

#test bump patch
[[ "$(vergo bump patch --tag-prefix=app --log-level=trace 2>&1)" == *'Set tag app-0.1.1'* ]]
[[ "$(vergo bump patch --tag-prefix=app --log-level=trace 2>&1)" == *'Push not enabled'* ]]
[[ "$(vergo bump patch --tag-prefix=app --log-level=error 2>&1)" == "0.1.1" ]]
#test bump patch check version
[[ "$(vergo get lr -t=app)" == "0.1.1" ]]
[[ "$(vergo get cv --tag-prefix=app)" == "0.1.1" ]]

#test bump major
readonly tagPrefixBananaVersion_2_0_0="$(vergo bump major --push-tag --tag-prefix=banana --log-level=trace 2>&1)"
[[ $tagPrefixBananaVersion_2_0_0 == *'Set tag banana-2.0.0'* ]]
[[ $tagPrefixBananaVersion_2_0_0 == *'Pushing tag: banana-2.0.0'* ]]
#test bump major check version
[[ "$(vergo get lr --tag-prefix=banana)" == "2.0.0" ]]
[[ "$(vergo get cv --tag-prefix=banana)" == "2.0.0" ]]

#test bump major ignoring strict host checking
readonly tagPrefixBananaVersion_3_0_0="$(vergo bump minor --push-tag --tag-prefix=mango --log-level=trace -d 2>&1)"
[[ $tagPrefixBananaVersion_3_0_0 == *'Set tag mango-0.1.0'* ]]
[[ $tagPrefixBananaVersion_3_0_0 == *'Pushing tag: mango-0.1.0'* ]]
#test bump major check version with strict host checking
[[ "$(vergo get lr --tag-prefix=mango)" == "0.1.0" ]]
[[ "$(vergo get cv --tag-prefix=mango)" == "0.1.0" ]]

#test bump major with slash postfix
readonly tagPrefixOrangeVersion_3_0_0="$(vergo bump major --push-tag --tag-prefix=orange/v --log-level=trace 2>&1)"
[[ $tagPrefixOrangeVersion_3_0_0 == *'Set tag orange/v3.0.0'* ]]
[[ $tagPrefixOrangeVersion_3_0_0 == *'Pushing tag: orange/v3.0.0'* ]]
#test bump major check version
[[ "$(vergo get lr --tag-prefix=orange/v)" == "3.0.0" ]]
[[ "$(vergo get cv --tag-prefix=orange/v)" == "3.0.0" ]]

#test list tags with tag-prefix=app
[[ "$(vergo list --tag-prefix=app --log-level=trace 2>&1)" == *'0.1.1
0.1.0' ]]
[[ "$(vergo list --tag-prefix=app --log-level=trace --sort-direction asc 2>&1)" == *'0.1.0
0.1.1' ]]
#test list tags with tag-prefix=apple
[[ "$(vergo list --tag-prefix=apple --log-level=trace 2>&1)" == *'0.2.0
0.1.1' ]]
[[ "$(vergo list --tag-prefix=apple --log-level=trace --sort-direction asc 2>&1)" == *'0.1.1
0.2.0' ]]

#test list tags with tag-prefix=orange
[[ "$(vergo list --tag-prefix=orange/v --log-level=trace 2>&1)" == *'3.0.0
2.3.0' ]]
[[ "$(vergo list --tag-prefix=orange/v --log-level=trace --sort-direction asc 2>&1)" == *'2.3.0
3.0.0' ]]

#test list tags with tag-prefix=app with prefix included
[[ "$(vergo list --tag-prefix=app -p --log-level=trace 2>&1)" == *'app-0.1.1
app-0.1.0' ]]
[[ "$(vergo list --tag-prefix=app -p --log-level=trace --sort-direction asc 2>&1)" == *'app-0.1.0
app-0.1.1' ]]

#test list tags with tag-prefix=apple with prefix included
[[ "$(vergo list --tag-prefix=apple -p --log-level=trace 2>&1)" == *'apple-0.2.0
apple-0.1.1' ]]
[[ "$(vergo list --tag-prefix=apple -p --log-level=trace --sort-direction asc 2>&1)" == *'apple-0.1.1
apple-0.2.0' ]]

#test push
[[ "$(vergo push --tag-prefix=apple --log-level=trace 2>&1)" == *'Pushing tag: apple-0.2.0'* ]]
[[ "$(vergo push --tag-prefix=apple --log-level=trace 2>&1)" == *'Pushing tag: apple-0.2.0'* ]]
[[ "$(vergo push --tag-prefix=apple --log-level=trace 2>&1)" == *'origin remote was up to date, no push done'* ]]
[[ "$(vergo push --tag-prefix=app --log-level=trace 2>&1)" == *'Pushing tag: app-0.1.1'* ]]
[[ "$(vergo push --tag-prefix=orange/v --log-level=trace 2>&1)" == *'Pushing tag: orange/v3.0.0'* ]]

#test push tags are present on remote
remote_tags=$(git ls-remote --tags origin apple-0.2.0 app-0.1.1 banana-2.0.0 orange/v3.0.0)
[[ "${remote_tags}" == *'refs/tags/app-0.1.1'* ]]
[[ "${remote_tags}" == *'refs/tags/apple-0.2.0'* ]]
[[ "${remote_tags}" == *'refs/tags/banana-2.0.0'* ]]
[[ "${remote_tags}" == *'refs/tags/orange/v3.0.0'* ]]

#test headless checkout
setUpLikeCI

[[ "$(git branch -l | grep -v HEAD)" == "" ]] #make sure no local branch
git checkout 117443bb
[[ "$(vergo check release --tag-prefix=apple 2>&1)" == *'invalid headless checkout'* ]]
[[ "$(vergo bump minor --tag-prefix=apple 2>&1)" == *'invalid headless checkout'* ]]

[[ "$(git branch -l | grep -v HEAD)" == "" ]] #make sure no local branch
git checkout a54f1f7
vergo check release --tag-prefix=apple
[[ "$(vergo bump minor --tag-prefix=apple --log-level=trace 2>&1)" == *'Set tag apple-0.2.0'* ]]

git checkout origin/master -b master #create local branch
[[ "$(vergo check release --tag-prefix=apple --log-level=trace --versioned-branch-names main 2>&1)" == *'branch master is not in main branches list: main'* ]]
[[ "$(vergo bump minor --tag-prefix=apple --log-level=trace --versioned-branch-names main 2>&1)" == *'branch master is not in main branches list: main'* ]]
