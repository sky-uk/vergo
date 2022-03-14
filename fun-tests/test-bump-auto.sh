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

some_temp_folder=$(mktemp -d)
cd "$some_temp_folder"
git init -q
git config user.email "vergo@example.com"
git config user.name "vergo"

TestAutoEmptyRepo() {
  [[ "$(vergo bump auto 2>&1)" == *'Error: reference not found'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'Error: reference not found'* ]]
  vergo check increment-hint                    #check is ok in an empty repo
  vergo check increment-hint --tag-prefix=apple #check is ok in an empty repo
} && TestAutoEmptyRepo

TestAutoNotMatchingCommitMessage() {
  touch some-file-1
  git add . && git commit -am"a commit"

  [[ "$(vergo bump auto 2>&1)" == *'Error: increment hint not present'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]
  [[ "$(vergo check increment-hint 2>&1)" == *'Error: increment hint not present'* ]]
  [[ "$(vergo check increment-hint --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]
} && TestAutoNotMatchingCommitMessage

TestAutoEmptyPrefix() {
  touch some-file-2
  git add . && git commit -am"[vergo:minor-release] a commit"

  [[ "$(vergo bump auto 2>&1)" == *'0.1.0'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]
  vergo check increment-hint 2>&1
  [[ "$(vergo check increment-hint --tag-prefix=apple 2>&1)" == *'Error: increment hint not present: apple'* ]]
} && TestAutoEmptyPrefix

TestAutoApplePrefix() {
  touch some-file-3
  git add . && git commit -am"[vergo:apple:major-release] a commit"

  [[ "$(vergo bump auto 2>&1)" == *'Error: increment hint not present'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'0.1.0'* ]]
  [[ "$(vergo check increment-hint 2>&1)" == *'Error: increment hint not present'* ]]
  vergo check increment-hint --tag-prefix=apple 2>&1

  touch some-file-4
  git add . && git commit -am"[vergo:apple:major-release] a commit"

  [[ "$(vergo bump auto 2>&1)" == *'Error: increment hint not present'* ]]
  [[ "$(vergo bump auto --tag-prefix=apple 2>&1)" == *'1.0.0'* ]]
  [[ "$(vergo check increment-hint 2>&1)" == *'Error: increment hint not present'* ]]
  vergo check increment-hint --tag-prefix=apple 2>&1

} && TestAutoApplePrefix
