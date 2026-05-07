#!/bin/bash

# Get the two most recent tags
read -r latest_tag n1_tag < <(git tag --sort=-version:refname | head -n 2 | tr '\n' ' ')

# Check if we have at least two tags
if [[ -z "${latest_tag:-}" ]] || [[ -z "${n1_tag:-}" ]]; then
    echo "Error: Not enough tags found. Need at least 2 tags." >&2
    exit 1
fi

# Get commits between the two tags, excluding those with [no-notes]
git log "${n1_tag}..${latest_tag}" --pretty=format:"- %s (%h)" --no-merges | \
    grep -v "\[no-notes\]" || true

echo ""
