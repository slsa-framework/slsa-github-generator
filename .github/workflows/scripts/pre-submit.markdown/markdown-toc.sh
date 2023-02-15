#!/bin/bash

# This script runs markdown-toc on Markdown files and detects if the table of
# contents has not been regenerated.

set -euo pipefail

# NOTE: Do not include issue templates since they contain Front Matter.
# markdown-toc will update Front Matter even if there is no TOC in the file.
# See: https://github.com/jonschlinkert/markdown-toc/issues/151
MD_FILES=$(
    find . -name '*.md' -type f \
        -not -iwholename '*/.git/*' \
        -not -iwholename '*/vendor/*' \
        -not -iwholename '*/node_modules/*' \
        -not -iwholename '*/.github/ISSUE_TEMPLATE/*'
)
for filename in ${MD_FILES}; do
    markdown-toc --bullets="-" -i "${filename}"
done

if [ "$(GIT_PAGER="cat" git diff --ignore-space-at-eol | wc -l)" -gt "0" ]; then
    echo "Detected TOC changes.  See status below:"
    GIT_PAGER="cat" git diff
    exit 1
fi
