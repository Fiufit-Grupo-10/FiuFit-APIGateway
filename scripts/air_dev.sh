#!/bin/bash -e
docker-compose build -f docker-compose.dev.yaml
docker-compose up -f docker-compose.dev.yaml web-dev
docker-compose down -f docker-compose.dev.yaml
