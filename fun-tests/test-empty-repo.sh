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

TestCurrentVersionEmptyRepo() (
  setup
  [[ "$(vergo get cv)" == "0.0.0-SNAPSHOT" ]]
  [[ "$(vergo get cv -p)" == "v0.0.0-SNAPSHOT" ]]
  [[ "$(vergo get cv -t apple)" == "0.0.0-SNAPSHOT" ]]
  [[ "$(vergo get cv -t apple -p)" == "apple-0.0.0-SNAPSHOT" ]]
  [[ "$(vergo bump minor 2>&1)" == "Error: reference not found" ]]
  [[ "$(vergo bump auto 2>&1)" == "Error: reference not found" ]]
)

TestCurrentVersionNoTags() (
  setup
  mktemp "$some_temp_folder/XXXXX"
  git add . && git commit -am"first commit"
  [[ "$(vergo get cv)" == "0.0.0-SNAPSHOT" ]]
  [[ "$(vergo get cv -p)" == "v0.0.0-SNAPSHOT" ]]
  [[ "$(vergo get cv -t apple)" == "0.0.0-SNAPSHOT" ]]
  [[ "$(vergo get cv -t apple -p)" == "apple-0.0.0-SNAPSHOT" ]]
)

tests=$(compgen -A function | grep -E '^Test')
for fn in $tests; do
  $fn
done
