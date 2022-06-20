#!/usr/bin/env bash

source "./.github/workflows/scripts/assert.sh"

e2e_assert_eq() {
    if ! assert_eq "$@"; then
        exit 1
    fi
}

e2e_assert_not_eq() {
    if ! assert_not_eq "$@"; then
        exit 1
    fi
}
