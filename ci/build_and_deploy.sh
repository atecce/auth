#!/bin/sh

GOOS=linux GOARCH=386 go build

gcloud compute scp auth atec@auth:~
gcloud compute ssh atec@auth --command="sudo mv /home/atec/auth /usr/sbin/"

gcloud compute ssh atec@auth --command="sudo systemctl restart auth.service"