-- name: init
CREATE TABLE IF NOT EXISTS beers (
	id              INTEGER PRIMARY KEY,
	bar_id          INTEGER NOT NULL,
	name            VARCHAR(256) NOT NULL,
	stock_quantity  INTEGER NOT NULL,
	purchase_price  DECIMAL(6, 2) NOT NULL,
	bottle_size     REAL NOT NULL,
	alcohol_content REAL NOT NULL,
	incr_coef       REAL NOT NULL,
	decr_coef       REAL NOT NULL,
	min_coef        REAL NOT NULL,
	max_coef        REAL NOT NULL
);

CREATE TABLE IF NOT EXISTS history (
	id            INTEGER PRIMARY KEY,
	beer_id       INTEGER NOT NULL,
	timestamp     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	sold_quantity INTEGER NOT NULL,
	selling_price DECIMAL(6, 2) NOT NULL,

	UNIQUE (beer_id, timestamp),
	FOREIGN KEY (beer_id) REFERENCES beers(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS users (
	id       INTEGER PRIMARY KEY,
	name     VARCHAR(256) UNIQUE NOT NULL,
	password VARCHAR(256) NOT NULL,
	admin    BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS tokens (
	value   VARCHAR(256) PRIMARY KEY,
	user_id INTEGER NOT NULL,

	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE IF NOT EXISTS entries (
	id       INTEGER PRIMARY KEY,
	timestamp     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	sold_quantity INTEGER NOT NULL,
	endOfParty    BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX IF NOT EXISTS users_name_index ON users(name);
