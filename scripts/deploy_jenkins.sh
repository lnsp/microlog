#!/bin/bash

docker-compose -f docker-compose.yml -f deployments/docker-compose.staging.yml -p microlog_staging up -d --build

echo "OK."
