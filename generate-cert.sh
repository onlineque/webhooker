#!/bin/bash

openssl req -new -newkey rsa:2048 -days 3650 -nodes -x509 -keyout selfsigned.key -out selfsigned.crt
