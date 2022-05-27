#!/usr/bin/env bash

source "./.github/workflows/scripts/assert.sh"

e2e_assert_eq() {
    assert_eq "$@"
    if [ "$?" != "0" ]; then
        exit 1
    fi
}

e2e_assert_not_eq() {
    assert_not_eq "$@"
    if [ "$?" != "0" ]; then
        exit 1
    fi
}
