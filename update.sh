#!/bin/bash

git pull 

soda migrate

go build -o bookings cmd/web/*.go

sudo supervisorctl stop booking
sudo supervisorctl start booking
sudo supervisorctl stop mailhog
sudo supervisorctl start mailhog