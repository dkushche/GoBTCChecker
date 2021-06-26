#!/bin/bash

http -v POST http://localhost:8080/user/create email=aa@gmail.com password=Aa12345678
sleep 2
http -v POST http://localhost:8080/user/create email=aa@gmail.com password=Aa12345678
sleep 2
