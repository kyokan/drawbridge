Note:

`start-drawbridge.sh` is empty because to prototype is meant to be run locally. Here is how to run it on your machine:

```

#!/usr/bin/env bash

./build/drawbridge --config ./local-config.yml \
    --identity-private-key 0dbbe8e4ae425a6d2687f1a7e3ba17bc98c673636790f1b8ad91193c05875ef1 \
    --private-key ae6ae8e5ccbfb04590405997ee2d52d2b330726137b875053c36d94e974d162f \
    --p2p-port 9736 \
    --rpc-port 8081 \
    --lnd-port 10101 \
    --lnd-host bob \
    --bootstrap-peers "127.0.0.1:9735|0x02ce7edc292d7b747fab2f23584bbafaffde5c8ff17cf689969614441e0527b900" \
    --database-url "postgres://postgres@localhost:5432/drawbridge_2?sslmode=disable"
```

This will point to the bitcoin infra running in docker.

You'll also need a postgres db called `drawbridge_2`.

To run the oter node, you can just do `make start`.