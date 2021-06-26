# GoBTCChecker

# How to run

```sh
make
./btcchecker
```

# How to test

```sh
http -v POST http://localhost:8080/user/create email=${VALID_MAIL} password=${VALID_PASSORD} # password length [6, 15] symbols
http -v --session=${SESSION} POST http://localhost:8080/user/login email=${VALID_MAIL} password=${VALID_PASSORD}
http -v --session=${SESSION} GET http://localhost:8080/btcRate
```

# How to configure

You may set `bind-address`, `log level`, `storage path` in existing configs/btcchecker.toml file or create your own and put its path as an argument to btcchecker(more ./btcchecker --help)
