BEGIN;
CREATE TABLE operations (
  seq         SERIAL          PRIMARY KEY,
  id          UUID            NOT NULL,
  namespace   VARCHAR(64)     NOT NULL,
  tx_id       UUID            NOT NULL,
  optype      VARCHAR(64)     NOT NULL,
  opstatus    VARCHAR(64)     NOT NULL,
  member   VARCHAR(1024),
  plugin      VARCHAR(64)     NOT NULL,
  backend_id  VARCHAR(256)    NOT NULL,
  created     BIGINT          NOT NULL,
  updated     BIGINT,
  error       VARCHAR         NOT NULL,
  info        BYTEA
);

CREATE UNIQUE INDEX operations_id ON operations(id);
CREATE INDEX operations_created ON operations(created);
CREATE INDEX operations_backend ON operations(backend_id);
CREATE INDEX operations_namespace ON operations(namespace);

COMMIT;