#!/usr/bin/env bash

set -e

# Move into the directory of repo_a and setup git user
git config user.name "GitHub Action"
git config user.email "action@github.com"

# Copy from api server to local
BRANCH_NAME="update-file-$(date +%Y%m%d%H%M%S)"
FILE_PATH=this/pkg/types_jsonschema.go
cp apiextensions-apiserver/pkg/apis/apiextensions/v1beta1/types_jsonschema.go this/pkg

# Check if there is a difference between the files
if git diff --exit-code "$FILE_PATH"; then
  echo "No changes detected, exiting."
  exit 0
fi

echo "Changes detected, creating a pull request..."

# Stage the changes
git add "$FILE_PATH"
git commit -m "Updated $FILE_PATH from repository B"

# Push the branch to repository A
git push origin "$BRANCH_NAME"

# Create a pull request using the GitHub CLI
gh auth login --with-token <<< "$GITHUB_TOKEN"
gh pr create --title "Sync $FILE_PATH from repo B" --body "This PR updates $FILE_PATH from repository B" --head "$BRANCH_NAME" --base main

echo "Pull request created successfully."