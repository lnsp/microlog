#!/bin/sh

STAGE=$1
TAG=$2

echo "Deploying to environment $STAGE ..."
cd $HOME/microlog
git fetch --all --tags

if [ "$STAGE" == "prod" ]; then
    PREFIX=microlog
    git checkout $TAG
else
    PREFIX=microlog_staging
    git checkout master
fi

docker-compose up -d -f deployments/docker-compose.yml -f deployments/docker-compose.$STAGE.yml -p $PREFIX --build

echo "OK."