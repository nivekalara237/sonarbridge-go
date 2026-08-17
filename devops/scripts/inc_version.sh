#!/bin/sh

set -e
increment_version() {
  version=$1
  type=$2

  major=$(echo "$version" | cut -d. -f1)
  minor=$(echo "$version" | cut -d. -f2)
  patch=$(echo "$version" | cut -d. -f3)

  case $type in
    "patch") patch=$(expr $patch + 1) ;;
    "minor") minor=$(expr $minor + 1); patch=0 ;;
    "major") major=$(expr $major + 1); minor=0; patch=0 ;;
  esac

  echo "$major.$minor.$patch"
}

if [ "$0" = "$BASH_SOURCE" ] || [ -z "$BASH_SOURCE" ]; then
  increment_version "$1" "$2"
fi