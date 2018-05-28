#!/bin/sh

cd "$1"
pwd
if [ -f ./.webhook/setup ]; then
    chmod +x ./.webhook/setup
    ./.webhook/setup
fi
