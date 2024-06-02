CREATE TABLE "coins" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying NOT NULL,
  "base" character varying NOT NULL
);

CREATE TABLE "exchanges" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "name" character varying NOT NULL,
  "url" character varying NOT NULL,
  "ticker_format" character varying NOT NULL
);

CREATE TABLE "exchange_coins" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "coin_id" uuid NOT NULL,
  "exchange_id" uuid NOT NULL,
  "name" character varying NOT NULL,
  "base" character varying NOT NULL
);

CREATE TABLE "exchange_tickers" (
  "id" uuid NOT NULL DEFAULT gen_random_uuid(),
  "exchange_id" uuid NOT NULL,
  "base_exch_coin_id" uuid NOT NULL,
  "quote_exch_coin_id" uuid NOT NULL
);

ALTER TABLE "coins"
ADD PRIMARY KEY ("id");

ALTER TABLE "exchanges"
ADD PRIMARY KEY ("id");

ALTER TABLE "exchange_coins"
ADD PRIMARY KEY ("id");

ALTER TABLE "exchange_tickers"
ADD PRIMARY KEY ("id");

ALTER TABLE "exchange_coins"
ADD CONSTRAINT "FK_COIN_ID"
FOREIGN KEY ("coin_id") REFERENCES "coins" ("id");

ALTER TABLE "exchange_coins"
ADD CONSTRAINT "FK_EXCHANGE_ID"
FOREIGN KEY ("exchange_id") REFERENCES "exchanges" ("id");

ALTER TABLE "exchange_tickers"
ADD CONSTRAINT "FK_EXCHANGE_ID"
FOREIGN KEY ("exchange_id") REFERENCES "exchanges" ("id");

ALTER TABLE "exchange_tickers"
ADD CONSTRAINT "FK_BASE_EXCH_COIN_ID"
FOREIGN KEY ("base_exch_coin_id") REFERENCES "exchange_coins" ("id");

ALTER TABLE "exchange_tickers"
ADD CONSTRAINT "FK_QUOTE_EXCH_COIN_ID"
FOREIGN KEY ("quote_exch_coin_id") REFERENCES "exchange_coins" ("id");

ALTER TABLE "exchange_coins"
ADD CONSTRAINT "coin_exchange_unique"
UNIQUE ("coin_id", "exchange_id");

ALTER TABLE "exchange_tickers"
ADD CONSTRAINT "exchange_base_quote_unique"
UNIQUE ("exchange_id", "base_exch_coin_id", "quote_exch_coin_id");

INSERT INTO "exchanges" ("id", "name", "url", "ticker_format") VALUES
('dd69994e-11bc-4c3e-b214-10bed360d404', 'Binance', 'https://www.binance.com', '%s%s'),
('df7fdcdd-e728-4a69-8f53-a6b339a514f9', 'Bitrue', 'https://www.bitrue.com', '%s_%s'),
('4f173f90-2cbb-4582-b725-e5abf58209e7', 'Gate', 'https://www.gate.io', '%s_%s'),
('19f2c852-a8ad-473a-a417-05d8bb4eed33', 'MEXC', 'https://www.mexc.com', '%s%s'),
('1e9b3253-3312-4407-a803-f51bcfa45933', 'XT', 'https://www.xt.com', '%s_%s');
