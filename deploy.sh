#!/bin/bash
cd /srv/babypro
git pull origin main
docker-compose down
docker-compose up --build -d
