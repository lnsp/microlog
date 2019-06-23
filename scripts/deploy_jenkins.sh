#!/bin/bash

docker-compose -f docker-compose.yml -f deployments/docker-compose.staging.yml -p staging up -d --build

echo "OK."
