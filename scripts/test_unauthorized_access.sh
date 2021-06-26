#!/bin/bash

SESSION=$RANDOM

http -v --session="$SESSION" GET http://localhost:8080/btcRate