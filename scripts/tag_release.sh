#!/bin/sh

RELEASE_TAG=$(date -u +"%Y-%m-%d-%H-%M")
echo Tagging release $RELEASE_TAG
git checkout master
git tag $RELEASE_TAG
git push origin $RELEASE_TAG
echo Published release $RELEASE_TAG