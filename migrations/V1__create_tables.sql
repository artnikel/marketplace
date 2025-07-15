CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	login TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL
);

CREATE TABLE items (
	id SERIAL PRIMARY KEY,
	title TEXT NOT NULL,
	description TEXT NOT NULL,
	image_url TEXT,
	price NUMERIC(10, 2) NOT NULL,
	author_id INTEGER NOT NULL,
	author_login TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT now()
);