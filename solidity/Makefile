.PHONY: test

compile:
	truffle compile

migrate:
	NODE_ENV=development truffle migrate --reset --network ganache

clean:
	rm -rf build

test:
	NODE_ENV=test truffle test

testnet:
	ganache-cli -m "candy maple cake sugar pudding cream honey rich smooth crumble sweet treat" -b 1 -v

develop:
	truffle develop --log