#!/bin/bash

# Bump version script for semantic versioning
# Usage: ./scripts/bump-version.sh [major|minor|patch]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Validate argument
if [[ -z "$1" ]]; then
    echo -e "${RED}Error: Please specify a version bump type${NC}"
    echo "Usage: $0 [major|minor|patch]"
    exit 1
fi

BUMP_TYPE="$1"

# Validate bump type
if [[ ! "$BUMP_TYPE" =~ ^(major|minor|patch)$ ]]; then
    echo -e "${RED}Error: Invalid bump type '${BUMP_TYPE}'${NC}"
    echo "Valid options: major, minor, patch"
    exit 1
fi

# Get the latest tag
LATEST_TAG=$(git tag -l --sort=-version:refname | head -1)

if [[ -z "$LATEST_TAG" ]]; then
    echo -e "${RED}Error: No tags found in repository${NC}"
    exit 1
fi

# Remove 'v' prefix if present
VERSION="${LATEST_TAG#v}"

# Parse version
if [[ ! "$VERSION" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
    echo -e "${RED}Error: Invalid version format '${VERSION}'${NC}"
    echo "Expected format: X.Y.Z"
    exit 1
fi

MAJOR="${BASH_REMATCH[1]}"
MINOR="${BASH_REMATCH[2]}"
PATCH="${BASH_REMATCH[3]}"

# Calculate new version
case "$BUMP_TYPE" in
    major)
        NEW_MAJOR=$((MAJOR + 1))
        NEW_MINOR=0
        NEW_PATCH=0
        ;;
    minor)
        NEW_MAJOR=$MAJOR
        NEW_MINOR=$((MINOR + 1))
        NEW_PATCH=0
        ;;
    patch)
        NEW_MAJOR=$MAJOR
        NEW_MINOR=$MINOR
        NEW_PATCH=$((PATCH + 1))
        ;;
esac

NEW_VERSION="${NEW_MAJOR}.${NEW_MINOR}.${NEW_PATCH}"
NEW_TAG="v${NEW_VERSION}"

# Display summary
echo -e "${YELLOW}Current version: ${GREEN}${LATEST_TAG}${NC}"
echo -e "${YELLOW}New version:     ${GREEN}${NEW_TAG}${NC}"
echo ""

# Confirm before proceeding
read -p "Create tag ${NEW_TAG}? (y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi

# Create tag
git tag "$NEW_TAG"
echo -e "${GREEN}✓ Tag ${NEW_TAG} created locally${NC}"

# Ask if user wants to push
echo ""
read -p "Push tag to remote? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git push origin "$NEW_TAG"
    echo -e "${GREEN}✓ Tag ${NEW_TAG} pushed to remote${NC}"
fi

echo ""
echo -e "${GREEN}Version bump complete!${NC}"
