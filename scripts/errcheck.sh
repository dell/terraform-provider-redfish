#!/usr/bin/env bash

# Check gofmt
echo "==> Checking for unchecked errors..."

if ! which errcheck > /dev/null; then
    echo "==> Installing errcheck..."
    go install github.com/kisielk/errcheck@latest
fi

errcheck -ignoretests -exclude errcheckexcludes.txt $(go list ./...| grep -v /vendor/)
