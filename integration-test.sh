#!/usr/bin/env bash

curl --request POST \
  --url http://localhost:8080/rpc \
  --header 'content-type: application/json' \
  --data '{
	"id": 1,
	"method": "FundingService.Approve",
	"params": [{
		"amount": "0x3E8"
	}]
}'

curl --request POST \
  --url http://localhost:8080/rpc \
  --header 'content-type: application/json' \
  --data '{
	"id": 1,
	"method": "FundingService.Deposit",
	"params": [{
		"amount": "0x3E8"
	}]
}'