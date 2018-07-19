CREATE TABLE channels (
  chain_id VARCHAR NOT NULL,
  finalized_id VARCHAR,
  temporary_id VARCHAR NOT NULL,
  funding_amount DECIMAL(72, 0) NOT NULL,
  push_amount DECIMAL(72, 0) NOT NULL,
  dust_limit DECIMAL(72, 0) NOT NULL,
  max_value_in_flight DECIMAL(72, 0),
  channel_reserve DECIMAL(72, 0),
  htlc_minimum DECIMAL(72, 0),
  fee_per_kw INTEGER,
  csv_delay INTEGER,
  max_accepted_htlcs INTEGER,
  our_funding_key VARCHAR,
  our_revocation_point VARCHAR,
  our_payment_point VARCHAR,
  our_delayed_payment_point VARCHAR,
  our_htlc_point VARCHAR,
  our_commitment_seed VARCHAR,
  their_funding_key VARCHAR,
  their_revocation_point VARCHAR,
  their_payment_point VARCHAR,
  their_delayed_payment_point VARCHAR,
  their_htlc_point VARCHAR,
  their_first_commitment_point VARCHAR,
  our_signature VARCHAR,
  their_signature VARCHAR,
  input_id VARCHAR REFERENCES eth_utxos(id),
  self_originated BOOLEAN
);

CREATE UNIQUE INDEX channels_finalized_id ON channels(finalized_id);
CREATE UNIQUE INDEX channels_temporary_id ON channels(temporary_id) WHERE finalized_id IS NULL;