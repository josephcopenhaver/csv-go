#!/bin/bash

main() {
    rm -f coverage.out
    go test -coverprofile=coverage.out $(go list ./... | grep -vE '^github\.com/josephcopenhaver/csv-go/internal/examples/')
    local n="$(go tool cover -func coverage.out | grep -E '^total:' | sed -E 's/^.*\s+([0-9\.]+)\s*%.*$/\1/ ; s/100.0*$/100/')"

    local color='red'
    if [[ $n == "100" ]]; then
        # color='green'
        color='rgb%2852%2C208%2C88%29'
    else
        n="$(printf %s "$n" | sed -E 's/\.[0-9]*$//')"
        if [[ $n -ge 90 ]]; then
            color='teal'
        elif [[ $n -ge 80 ]]; then
            color='blue'
        elif [[ $n -ge 60 ]]; then
            color='orange'
        else
            color='red'
        fi
    fi

    rm -f README.md.bak
    sed -i.bak -E 's/^!\[code-coverage\]\(.*\)\s*$/![code-coverage](https:\/\/img.shields.io\/badge\/code_coverage-'"${n}"'%25-'"${color}"')/g' README.md
    rm README.md.bak
}


(set -euo pipefail ; main "$@")
