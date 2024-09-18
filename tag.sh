#!/usr/bin/env bash
set -eu

# Extract the latest version from CHANGELOG.md
VERSION=$(sed -n 's/^## \[\([0-9.]*\)\].*/\1/p' CHANGELOG.md | head -n 1)

# Generate the tag
echo "==> Tagging version $VERSION..."
echo git tag -a "v${VERSION}" -m "Version $VERSION"
git tag -a "v${VERSION}" -m "Version $VERSION"
git push --tags

exit 0