#!/bin/bash

# This script runs markdown-toc on Markdown files and detects if the table of
# contents has not been regenerated.

set -euo pipefail

if [ "$(GIT_PAGER="cat" git diff --ignore-space-at-eol | wc -l)" -gt "0" ]; then
    echo "Detected TOC changes.  See status below:"
    GIT_PAGER="cat" git diff
    exit 1
fi
