package main

import (
    "testing"
    "context"
    "log"
    "sync"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "bytes"
)

func BenchmarkImportLibrary(b *testing.B) {
    var err error
	server := Server{
        wg: sync.WaitGroup{},
        cache: make(map[string]struct{}),
        templates: make(map[string]string),
        urls: make(map[string]string),
        ctx: context.Background(),
        artistCache: sync.Map{},
        artistCreditCache: sync.Map{},
        artistCreditNameCache: sync.Map{},
        batch : &pgx.Batch{},
        bufferPool: sync.Pool{
            New: func() interface{} {
                return new(bytes.Buffer)
            },
        },
        ffprobePool: sync.Pool{
            New: func() interface{} {
                return Ffprobe{}
            },
        },
    }
    server.ctx = context.Background()
    config, err := pgxpool.ParseConfig("postgres://localhost:5432/mbms?pool_max_conns=100")
    server.db, err = pgxpool.NewWithConfig(server.ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer server.db.Close()
	server.batch.Queue(" DROP SCHEMA IF EXISTS mbms CASCADE") 
	server.batch.Queue("CREATE SCHEMA mbms")
    server.batch.Queue("CREATE TABLE mbms.artist (mbid VARCHAR(36) NOT NULL, name VARCHAR, unique (mbid))")
    server.batch.Queue("CREATE TABLE mbms.artist_credit (name VARCHAR NOT NULL, id INTEGER NOT NULL, unique (name))")
    server.batch.Queue("CREATE TABLE mbms.artist_credit_name (artist_credit INTEGER NOT NULL, name VARCHAR NOT NULL, artist_mbid VARCHAR(36) NOT NULL, unique (artist_credit, artist_mbid))")
    server.batch.Queue("CREATE TABLE mbms.release (mbid VARCHAR(36) NOT NULL, name VARCHAR NOT NULL, artist_credit INTEGER NOT NULL, date VARCHAR, cover_url VARCHAR, unique (mbid))")
    server.batch.Queue("CREATE TABLE mbms.track (mbid VARCHAR(36) NOT NULL, name VARCHAR NOT NULL, artist_credit INTEGER NOT NULL, release_mbid VARCHAR(36) NOT NULL, number INTEGER NOT NULL, length INTEGER NOT NULL, url VARCHAR, unique (mbid))")
    server.batch.Queue("CREATE TABLE mbms.playlist (id SERIAL PRIMARY KEY, name VARCHAR NOT NULL, unique (name))")
    server.batch.Queue("CREATE TABLE mbms.playlist_track (playlist INTEGER NOT NULL, track VARCHAR(36) NOT NULL, unique (playlist, track))")
    server.batch.Queue("SET search_path TO mbms")
    err = server.db.SendBatch(server.ctx, server.batch).Close()
    if err != nil {
        log.Fatal(err)
    }
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        server.wg.Add(1)
        go server.ImportLibrary("/backup/encoded")
        server.wg.Wait()
    }
}
