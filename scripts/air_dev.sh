#!/bin/bash -e
docker-compose build
docker-compose up web-dev
docker-compose down
