#! /bin/sh
set -e

sudo chown gouser:gouser /go-cache
devspace run setup
devspace run generate

exec "$@"
