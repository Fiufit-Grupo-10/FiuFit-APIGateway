#!/bin/bash -e
echo $(pwd)
docker-compose -f docker-compose.dev.yml build 
docker-compose -f docker-compose.dev.yml up web-dev
docker-compose -f docker-compose.dev.yml down
