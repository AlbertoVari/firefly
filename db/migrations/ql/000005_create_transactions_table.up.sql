CREATE TABLE transactions (
  id          string   NOT NULL,
  ttype       string   NOT NULL,
  namespace   string   NOT NULL,
  ref         string,
  signer      string   NOT NULL,
  hash        string   NOT NULL,
  created     int64    NOT NULL,
  protocol_id string,
  status      string   NOT NULL,
  info        blob
);

CREATE UNIQUE INDEX transactions_primary ON transactions(id);
CREATE INDEX transactions_created ON transactions(created);
CREATE INDEX transactions_protocol_id ON transactions(protocol_id);
CREATE INDEX transactions_ref ON transactions(ref);

