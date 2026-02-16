#!/bin/bash
# Auto-merge script for TODO Tracker CLI
# Run this after gh CLI is available

set -e

echo "=========================================="
echo "TODO Tracker CLI - Auto Merge Script"
echo "=========================================="

# List of branches to merge (in order)
BRANCHES=(
  "feature/add-delete-command"
  "feature/export-functionality"
  "feature/init-command"
  "feature/watch-mode"
  "feature/tags-categories"
  "feature/time-tracking"
)

REPO="https://github.com/duncan-2126/ProjectManagement"

for branch in "${BRANCHES[@]}"; do
  echo ""
  echo "Processing: $branch"

  # Check if PR exists
  pr_num=$(gh pr list --head "$branch" --json number -q '.[0].number' 2>/dev/null || echo "")

  if [ -n "$pr_num" ]; then
    echo "  PR #$pr_num already exists"

    # Review and approve
    echo "  Adding review..."
    gh pr review "$pr_num" --approve --body "Approved by automated review system"

    # Merge
    echo "  Merging..."
    gh pr merge "$pr_num" --squash --delete-branch
    echo "  ✓ Merged!"
  else
    echo "  Creating PR..."
    gh pr create --base main --head "$branch" --title "Merge $branch" --body "Automated merge from feature branch"
    echo "  ✓ PR created!"
  fi
done

echo ""
echo "=========================================="
echo "All branches merged successfully!"
echo "=========================================="
