CREATE TABLE eth_chainsaw_status (
  last_seen_block BIGINT,
  last_polled_at BIGINT
);

INSERT INTO eth_chainsaw_status (last_seen_block, last_polled_at) VALUES (0, 0);