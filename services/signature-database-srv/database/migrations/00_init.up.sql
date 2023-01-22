CREATE TABLE fourbyte
(
    name varchar PRIMARY KEY,
    hash bytea
);

CREATE INDEX IF NOT EXISTS fourbyte_hash ON fourbyte USING btree (hash);

CREATE TABLE thirtytwobyte
(
    name varchar PRIMARY KEY,
    hash bytea
);

CREATE INDEX IF NOT EXISTS thirtytwobyte_hash ON thirtytwobyte USING btree (hash);
