#!/bin/bash

http -v POST http://localhost:8000/user/create email=dima.kushhevskij@gmail.com password=Aa12345678
sleep 2
http -v --session=user POST http://localhost:8080/user/login email=dima.kushhevskij@gmail.com password=Aa12345678
sleep 2
http -v --session=user GET http://localhost:8080/btcRate
