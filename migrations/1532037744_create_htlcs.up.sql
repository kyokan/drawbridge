CREATE TABLE htlcs (
  id INT NOT NULL,
  channel_id VARCHAR NOT NULL REFERENCES channels(finalized_id),
  amount DECIMAL(72, 0),
  payment_hash VARCHAR NOT NULL,
  cltv_expiry INT NOT NULL,
  onion_routing_packet VARCHAR NOT NULL,
  preimage VARCHAR,
  self_originated BOOLEAN NOT NULL
);

CREATE UNIQUE INDEX htlcs_id_channel_id_index ON htlcs(id, channel_id);