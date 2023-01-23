R"###(
CREATE TABLE artist (
    mbid                VARCHAR(36) NOT NULL,
    name                VARCHAR,
    unique (mbid)
);

CREATE TABLE artist_credit (
    name                VARCHAR NOT NULL,
    unique (name)
);
CREATE TABLE artist_credit_name (
    artist_credit       INTEGER NOT NULL,
    artist              INTEGER NOT NULL,
    name                VARCHAR NOT NULL,
    unique (artist_credit, artist)
);

CREATE TABLE track (
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
    mbid                VARCHAR(36) NOT NULL,
    name                VARCHAR NOT NULL,
    artist_credit       INTEGER NOT NULL,
    date                VARCHAR,
    type                INTEGER NOT NULL,
    cover_url           VARCHAR,
    unique (mbid)
);
CREATE TABLE type (
    name                VARCHAR NOT NULL,
    unique (name)
);
CREATE TABLE l_artist_artist (
    artist             INTEGER NOT NULL,
    l_artist           INTEGER NOT NULL,
    type               INTEGER NOT NULL,
    direction          INTEGER NOT NULL,
    unique (artist, l_artist)
);
CREATE TABLE t_artist_artist (
    name               VARCHAR NOT NULL,
    unique (name)
);
CREATE TABLE playlist (
    name               VARCHAR NOT NULL,
    comment            VARCHAR,
    unique (name)
);
CREATE TABLE playlist_track (
    playlist           INTEGER NOT NULL,
    track              INTEGER NOT NULL
);
)###"
