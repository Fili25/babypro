#!/bin/bash
cd /root/babytracker
git pull origin main
docker-compose down
docker-compose up --build -d
