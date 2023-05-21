#include <sqlite3.h>
#include "Types.h"
#pragma once

class DatabaseManager
{
public:
    DatabaseManager(char const* database_path);
    ~DatabaseManager();
    int add_artist(Artist const* artist);
    int add_track(Track const* track);
    int add_release(Release const* release);
    int add_release_type(Type const* type);
    int add_artist_credit(ArtistCredit const* artist_credit);
    int add_artist_credit_name(ArtistCreditName const* artist_credit_name);
    int add_t_artist_artist(char const* name);
    int get_release_id(char const* mbid);
    int get_artist_credit_id(char const* name);
    int get_artist_id(char const* mbid);
    int get_t_artist_artist_id(char const* name);
    int get_release(char const* mbid, std::string &name);
    bool update_artist_info();

private:
    sqlite3* m_database;
    const char* m_database_path;
    sqlite3_stmt* m_statement;
    char const* m_query;
    int m_last_artist;
};



