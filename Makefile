compile-contracts:
	@$(MAKE) -C ./solidity compile

migrate-contracts:
	@$(MAKE) -C ./solidity migrate

migrate-database:
	migrate -database "$(DATABASE_URL)" -path ./migrations up

create-db-migration:
	cd  ./migrations && migrate create -ext sql $(MIGRATION_NAME)

develop:
	@$(MAKE) -C ./solidity testnet

test: compile
	@$(MAKE) -C ./solidity test
	go test ./pkg/...
	go test ./internal/...

compile-extract-abi:
	go build -o build/extract-abi ./cmd/extract_abi.go

abigen: compile-contracts compile-extract-abi
	mkdir -p ./build/abi
	./build/extract-abi --contracts ./solidity/build/contracts/UTXOToken.json,./solidity/build/contracts/ERC20.json --output-dir ./build/abi
	abigen --abi ./build/abi/UTXOToken.json --pkg contracts --type UTXOToken --out ./pkg/contracts/utxo_token.go
	abigen --abi ./build/abi/ERC20.json --pkg contracts --type ERC20 --out ./pkg/contracts/erc20.go

compile: abigen
	go build -o ./build/drawbridge ./cmd/drawbridge.go

dep:
	dep ensure -v
	cp -r \
      "${GOPATH}/src/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1" \
      "vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/"

clean:
	@$(MAKE) -C ./solidity clean
	rm -rf ./build
	rm -rf ./pkg/contracts/*.go

start: compile
	./build/drawbridge --config ./local-config.yml