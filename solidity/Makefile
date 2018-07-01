.PHONY: test

compile:
	truffle compile

migrate:
	NODE_ENV=development truffle migrate --reset

clean:
	rm -rf build

test:
	NODE_ENV=test truffle test

testnet:
	truffle develop --log