#!/bin/bash

SESSION=$RANDOM

http -v --session="$SESSION" POST http://localhost:8080/user/login email=bbb@gmail.com password=Aa12345678
