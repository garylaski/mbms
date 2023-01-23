#include "DatabaseManager.h"
#include <sqlite3.h>
#include <cstddef>
#include <stdexcept>
#include <iostream>
#include <musicbrainz5/Query.h>
#include <musicbrainz5/Artist.h>
#include <musicbrainz5/RelationListList.h>
#include <musicbrainz5/RelationList.h>
#include <musicbrainz5/Relation.h>

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
int DatabaseManager::get_t_artist_artist_id(char const* name) {
    m_query = "SELECT rowid FROM t_artist_artist WHERE name = ?";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, name, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    int id = sqlite3_column_int(m_statement, 0);
    sqlite3_finalize(m_statement);
    return id;
}
int DatabaseManager::add_t_artist_artist(char const* name) {
    m_query = "INSERT INTO t_artist_artist (name) VALUES (?)";
    sqlite3_prepare_v2(m_database, m_query, -1, &m_statement, NULL);
    sqlite3_bind_text(m_statement, 1, name, -1, SQLITE_STATIC);
    sqlite3_step(m_statement);
    sqlite3_finalize(m_statement);
    return sqlite3_last_insert_rowid(m_database);
}
bool DatabaseManager::update_artist_info() {
    // select all artists and loop over them
    m_query = "SELECT rowid, mbid FROM artist";
    sqlite3_stmt *statement, *statement2;
    sqlite3_prepare_v2(m_database, m_query, -1, &statement2, NULL);
    MusicBrainz5::CQuery Query("ArtistName");
    MusicBrainz5::CMetadata metadata;
    MusicBrainz5::CQuery::tParamMap params;
    params["inc"] = "artist-rels";
    while (sqlite3_step(statement2) == SQLITE_ROW) {
        int id = sqlite3_column_int(statement2, 0);
        char const* mbid = (char const*)sqlite3_column_text(statement2, 1);
        metadata = Query.Query("artist", mbid, "", params);
        // loop over all artist relations
        int direction;
        if (metadata.Artist()->RelationListList()) {
            for (int i = 0; i < metadata.Artist()->RelationListList()->NumItems(); i++) {
                MusicBrainz5::CRelationList* relation_list = metadata.Artist()->RelationListList()->Item(i);
                for (int j = 0; j < relation_list->NumItems(); j++) {
                    MusicBrainz5::CRelation* relation = relation_list->Item(j);
                    direction = relation->Direction() == "forward" ? 1 : 0;
                    // check if the artist exists
                    int l_id = get_artist_id(relation->Target().c_str());
                    if (l_id == 0) {
                        continue;
                    }
                    // get the relation type id
                    int type_id = get_t_artist_artist_id(relation->Type().c_str());
                    if (type_id == 0) {
                        type_id = add_t_artist_artist(relation->Type().c_str());
                    }
                    m_query = "INSERT INTO l_artist_artist (artist, l_artist, type, direction) VALUES (?, ?, ?, ?)";
                    sqlite3_prepare_v2(m_database, m_query, -1, &statement, NULL);
                    sqlite3_bind_int(statement, 1, id);
                    sqlite3_bind_int(statement, 2, l_id);
                    sqlite3_bind_int(statement, 3, type_id);
                    sqlite3_bind_int(statement, 4, direction);
                    int rc = sqlite3_step(statement);
                    if (rc != SQLITE_DONE) {
                        std::cout << "Error inserting l_artist_artist: " << sqlite3_errmsg(m_database) << std::endl;
                        return false;
                    }
                    sqlite3_finalize(statement);
                }
            }
        }
        char const* name = metadata.Artist()->Name().c_str();
        // update artist info
        m_query = "UPDATE artist SET name = ? WHERE rowid = ?";
        sqlite3_prepare_v2(m_database, m_query, -1, &statement, NULL);
        sqlite3_bind_text(statement, 1, name, -1, SQLITE_STATIC);
        sqlite3_bind_int(statement, 2, id);
        sqlite3_step(statement);
        sqlite3_finalize(statement);
    }
    sqlite3_finalize(statement2);
    return true;
}
