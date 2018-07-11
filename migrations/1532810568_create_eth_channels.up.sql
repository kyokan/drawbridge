CREATE TABLE eth_channels (
  id VARCHAR NOT NULL PRIMARY KEY,
  funding_output VARCHAR NOT NULL REFERENCES eth_outputs(id),
  counterparty VARCHAR NOT NULL
);