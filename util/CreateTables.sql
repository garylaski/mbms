CREATE TABLE artist (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    mbid                VARCHAR(36) NOT NULL,
    name                VARCHAR,
    unique (mbid)
);

CREATE TABLE artist_credit (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    name                VARCHAR NOT NULL,
    artist_count        INTEGER NOT NULL,
    unique (name)
);

CREATE TABLE artist_credit_name (
    artist_credit       INTEGER NOT NULL,
    position            INTEGER NOT NULL,
    artist              INTEGER NOT NULL,
    name                VARCHAR NOT NULL,
    unique (artist_credit, position)
);

CREATE TABLE track (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    mbid                VARCHAR(36) NOT NULL,
    name                VARCHAR NOT NULL,
    number              INTEGER NOT NULL,
    artist_credit       INTEGER NOT NULL,
    length              INTEGER NOT NULL,
    release             INTEGER NOT NULL,
    url                 VARCHAR,
    unique (mbid)
);

CREATE TABLE release (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    mbid                VARCHAR(36) NOT NULL,
    name                VARCHAR NOT NULL,
    artist_credit       INTEGER NOT NULL,
    date                INTEGER NOT NULL,
    type                INTEGER NOT NULL,
    cover_url           VARCHAR,
    unique (mbid)
);
CREATE TABLE type (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    name                VARCHAR NOT NULL,
    unique (name)
);
