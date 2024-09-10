#!/bin/sh
docker build --platform="darwin/arm64" -t ogglord/comp2unraid .
docker build --platform="linux/amd64" -t ogglord/comp2unraid .
docker build --platform="linux/arm64" -t ogglord/comp2unraid .