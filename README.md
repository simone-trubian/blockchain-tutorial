# The Blockchain Tutorial

## Install
```
dep ensure
go install ./cmd/...
```

## Usage
### List all possible commands
```
sb help
```

### Run sb blockchain
```
sb run --datadir=~/.sb
```

### Create a new account
```
sb wallet new-account --datadir=~/.sb 
```

## HTTP Usage
### List all balances
```
curl -X GET http://localhost:8080/balances/list -H 'Content-Type: application/json'
```

### Send and sign a new TX
```
curl --location --request POST 'http://localhost:8080/tx/add' \
--header 'Content-Type: application/json' \
--data-raw '{
	"from": "0x22ba1f80452e6220c7cc6ea2d1e3eeddac5f694a",
	"from_pwd": "security123",
	"to": "0x6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8",
	"value": 100
}'
```

## Compile
To local OS:
```
go install ./cmd/...
```

To cross-compile:
```
xgo --targets=linux/amd64 ./cmd/sb
```

## Tests
Run all tests with verbosity but one at a time, without timeout, to avoid ports collisions:
```
go test -v -p=1 -timeout=0 ./...
```

**Note:** The majority of tests are integration tests and take time. Expect the test suite to finish in ~30 mins. 