#!/bin/bash
rsync -avz --exclude 'node_modules' --exclude '.git' --exclude 'frontend/dist' -e "ssh -o StrictHostKeyChecking=no" /home/kali/Desktop/projects/Django/Kafe/ root@157.180.118.115:/var/www/kafe/
