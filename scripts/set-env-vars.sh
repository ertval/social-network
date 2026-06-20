#!/usr/bin/env bash
set -euo pipefail

export MY_MESSAGE="Hello World"
export MY_NUM=100
export MY_PI=3.142

printenv | grep -E '^(MY_MESSAGE|MY_NUM|MY_PI)='
