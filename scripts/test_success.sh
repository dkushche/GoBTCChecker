#!/bin/bash

SESSION=$RANDOM

http -v POST http://localhost:8080/user/create email=dima.kushhevskij@gmail.com password=Aa12345678
sleep 2
http -v --session="$SESSION" POST http://localhost:8080/user/login email=dima.kushhevskij@gmail.com password=Aa12345678
sleep 2
http -v --session="$SESSION" GET http://localhost:8080/btcRate
