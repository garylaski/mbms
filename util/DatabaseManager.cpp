#include "DatabaseManager.h"
#include <sqlite3.h>
#include <cstddef>
#include <stdexcept>
#include <iostream>

DatabaseManager::DatabaseManager(char const* database_path): 
    m_database_path(database_path) 
{
    if (sqlite3_open(":memory:", &m_database) != SQLITE_OK) {
        throw std::runtime_error("Failed to open database");
    }
    char const* schema = {
#include "Schema.sql"
    };
    if (sqlite3_exec(m_database, schema, NULL, NULL, NULL) != SQLITE_OK) {
        throw std::runtime_error("Failed to create database schema");
    }
}
DatabaseManager::~DatabaseManager() {
    sqlite3* file;
    sqlite3_open(m_database_path, &file);
    sqlite3_backup* backup = sqlite3_backup_init(file, "main", m_database, "main");
    sqlite3_backup_step(backup, -1);
    sqlite3_backup_finish(backup);
    sqlite3_close(m_database);
    sqlite3_close(file);
}
int DatabaseManager::add_artist(Artist const* artist) {
    m_query = "INSERT INTO artist (mbid, name) VALUES (?, ?) ON CONFLICT(mbid) DO UPDATE SET name = excluded.name WHERE excluded.name != NULL";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, artist->mbid, -1, SQLITE_STATIC);
    sqlite3_bind_text(m_statement, 2, artist->name, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    sqlite3_finalize(m_statement);
    return sqlite3_last_insert_rowid(m_database);
}
int DatabaseManager::add_track(Track const* track) {
    m_query = "INSERT INTO track (mbid, name, url, number, artist_credit, length, release) VALUES (?, ?, ?, ?, ?, ?, ?)";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, track->mbid, -1, SQLITE_STATIC);
    sqlite3_bind_text(m_statement, 2, track->name, -1, SQLITE_STATIC);
    sqlite3_bind_text(m_statement, 3, track->url, -1, SQLITE_STATIC);
    sqlite3_bind_int(m_statement, 4, track->number);
    sqlite3_bind_int(m_statement, 5, track->artist_credit);
    sqlite3_bind_int(m_statement, 6, track->length);
    sqlite3_bind_int(m_statement, 7, track->release);
    sqlite3_step(m_statement);
    sqlite3_finalize(m_statement);
    return sqlite3_last_insert_rowid(m_database);
}
int DatabaseManager::add_release(Release const* release) {
    m_query = "INSERT INTO release (mbid, name, date, cover_url, artist_credit, type) VALUES (?, ?, ?, ?, ?, ?)";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, release->mbid, -1, SQLITE_STATIC);
    sqlite3_bind_text(m_statement, 2, release->name, -1, SQLITE_STATIC);
    sqlite3_bind_text(m_statement, 3, release->date, -1, SQLITE_STATIC);
    sqlite3_bind_text(m_statement, 4, release->cover_url, -1, SQLITE_STATIC);
    sqlite3_bind_int(m_statement, 5, release->artist_credit);
    sqlite3_bind_int(m_statement, 6, release->type);
    sqlite3_step(m_statement);
    sqlite3_finalize(m_statement);
    return sqlite3_last_insert_rowid(m_database);
}
int DatabaseManager::add_release_type(Type const* type) {
    m_query = "INSERT INTO type (name) VALUES (?)";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, type->name, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    sqlite3_finalize(m_statement);
    return sqlite3_last_insert_rowid(m_database);
}
int DatabaseManager::add_artist_credit(ArtistCredit const* artist_credit) {
    m_query = "INSERT INTO artist_credit (name) VALUES (?)";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, artist_credit->name, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    sqlite3_finalize(m_statement);
    return sqlite3_last_insert_rowid(m_database);
}
int DatabaseManager::add_artist_credit_name(ArtistCreditName const* artist_credit_name) {
    m_query = "INSERT INTO artist_credit_name (artist_credit, artist, name) VALUES (?, ?, ?)";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_int(m_statement, 1, artist_credit_name->artist_credit);
    sqlite3_bind_int(m_statement, 2, artist_credit_name->artist);
    sqlite3_bind_text(m_statement, 3, artist_credit_name->name, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    sqlite3_finalize(m_statement);
    return sqlite3_last_insert_rowid(m_database);
}
int DatabaseManager::get_release_id(char const* mbid) {
    m_query = "SELECT rowid FROM release WHERE mbid = ?";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, mbid, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    int id = sqlite3_column_int(m_statement, 0);
    sqlite3_finalize(m_statement);
    return id;
}
int DatabaseManager::get_artist_credit_id(char const* name) {
    m_query = "SELECT rowid FROM artist_credit WHERE name = ?";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, name, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    int id = sqlite3_column_int(m_statement, 0);
    sqlite3_finalize(m_statement);
    return id;
}
int DatabaseManager::get_artist_id(char const* mbid) {
    m_query = "SELECT rowid FROM artist WHERE mbid = ?";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, mbid, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    int id = sqlite3_column_int(m_statement, 0);
    sqlite3_finalize(m_statement);
    return id;
}
