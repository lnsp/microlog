#!/bin/bash

STAGE=$1
TAG=$2

echo "Deploying to environment $STAGE ..."
cd $HOME/microlog
git fetch --all --tags

if [ "$STAGE" == "prod" ]; then
    PREFIX=microlog
    git checkout $TAG
    git pull
else
    PREFIX=microlog_staging
    git checkout master
    git pull
fi

docker-compose -f docker-compose.yml -f deployments/docker-compose.$STAGE.yml -p $PREFIX up -d --build

echo "OK."
