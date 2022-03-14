#!/usr/bin/env bash

set -euox pipefail
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

setup() {
  some_temp_folder=$(mktemp -d)
  cd "$some_temp_folder"
  git init -q
  git config user.email "vergo@example.com"
  git config user.name "vergo"
}

TestAutoEmptyRepo() (
  setup
  [[ "$(vergo bump auto 2>&1)" == *'Error: reference not found'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'Error: reference not found'* ]]
  vergo check increment-hint                    #check is ok in an empty repo
  vergo check increment-hint --tag-prefix=apple #check is ok in an empty repo
)

TestAutoNotMatchingCommitMessage() (
  setup
  mktemp "$some_temp_folder/XXXXX"
  git add . && git commit -am"a commit"

  [[ "$(vergo bump auto 2>&1)" == *'Error: increment hint not present'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]
  [[ "$(vergo check increment-hint 2>&1)" == *'Error: increment hint not present'* ]]
  [[ "$(vergo check increment-hint --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]
)

TestAutoEmptyPrefix() (
  setup
  mktemp "$some_temp_folder/XXXXX"
  git add . && git commit -am"[vergo:minor-release] a commit"

  [[ "$(vergo bump auto 2>&1)" == *'0.1.0'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]
  vergo check increment-hint 2>&1
  [[ "$(vergo check increment-hint --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]
)

TestAutoApplePrefix() (
  setup
  mktemp "$some_temp_folder/XXXXX"
  git add . && git commit -am"[vergo:apple:major-release] a commit"

  [[ "$(vergo bump auto 2>&1)" == *'Error: increment hint not present'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'0.1.0'* ]]
  [[ "$(vergo check increment-hint 2>&1)" == *'Error: increment hint not present'* ]]
  vergo check increment-hint --tag-prefix=apple 2>&1

  mktemp "$some_temp_folder/XXXXX"
  git add . && git commit -am"[vergo:apple:major-release] a commit"

  [[ "$(vergo bump auto 2>&1)" == *'Error: increment hint not present'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'1.0.0'* ]]
  [[ "$(vergo check increment-hint 2>&1)" == *'Error: increment hint not present'* ]]
  vergo check increment-hint --tag-prefix=apple 2>&1
)

TestCheckIncrementWithSkipReleaseHint() (
  setup
  mktemp "$some_temp_folder/XXXXX"
  git add . && git commit -am"[vergo:skip-release] a commit"
  vergo check increment-hint 2>&1
  [[ "$(vergo check increment-hint --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]

  mktemp "$some_temp_folder/XXXXX"
  git add . && git commit -am"[vergo:apple:skip-release] a commit"
  [[ "$(vergo check increment-hint 2>&1)" == *'Error: increment hint not present'* ]]
  vergo check increment-hint --tag-prefix=apple 2>&1
)

tests=$(compgen -A function | grep -E '^Test')
for fn in $tests; do
  $fn
done
