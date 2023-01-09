#include <taglib/tag.h>
#include <taglib/fileref.h>
#include <taglib/tpropertymap.h>
#include <sqlite3.h>
#include <stdio.h>
#include <fstream>
#include <string>
#include <filesystem>

std::string audio_ext = "mp3flacggwmaaacopuswavm4a";
struct artist_credit_name {
    int artist_credit;
    int position;
    int artist;
    std::string name;
};
struct track {
    int id;
    std::string mbid;
    std::string name;
    int number;
    int artist_credit;
    int length;
    int release;
    std::string url;
};
struct release {
    int id;
    std::string mbid;
    std::string name;
    int artist_credit;
    int date;
    int type;
    std::string cover_url;
};

int get_artist_id(sqlite3 *db, std::string artist_mbid) {
    std::string query = "SELECT id FROM artist WHERE mbid = '" + artist_mbid + "'";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_step(stmt);
    int artist_id = sqlite3_column_int(stmt, 0);
    return artist_id;
}

int add_artist(sqlite3 *db, std::string artist_mbid) {
    int artist_id = get_artist_id(db, artist_mbid);
    if (artist_id != 0) {
        return artist_id;
    }
    std::string query = "INSERT INTO artist (mbid) VALUES (?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, artist_mbid.c_str(), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    artist_id = sqlite3_last_insert_rowid(db);
    return artist_id;
}
int get_artist_credit_id(sqlite3 *db, std::string artist_credit, int artist_count) {
    std::string query = "SELECT id FROM artist_credit WHERE name = '" + artist_credit + "' AND artist_count = " + std::to_string(artist_count);
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_step(stmt);
    int artist_credit_id = sqlite3_column_int(stmt, 0);
    return artist_credit_id;
}

int add_artist_credit(sqlite3 *db, std::string artist_credit, int artist_count) {
    int artist_credit_id = get_artist_credit_id(db, artist_credit, artist_count);
    if (artist_credit_id != 0) {
        return artist_credit_id;
    }
    std::string query = "INSERT INTO artist_credit (name, artist_count) VALUES (?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, artist_credit.c_str(), -1, SQLITE_STATIC);
    sqlite3_bind_int(stmt, 2, artist_count);
    sqlite3_step(stmt);
    artist_credit_id = sqlite3_last_insert_rowid(db);
    return artist_credit_id;
}


int get_artist_credit_name_id(sqlite3 *db, int artist_credit, int position, int artist, std::string name) {
    std::string query = "SELECT id FROM artist_credit_name WHERE artist_credit = " + std::to_string(artist_credit) + " AND position = " + std::to_string(position) + " AND artist = " + std::to_string(artist) + " AND name = '" + name + "'";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_step(stmt);
    int artist_credit_name_id = sqlite3_column_int(stmt, 0);
    return artist_credit_name_id;
}
int add_artist_credit_name(sqlite3 *db, artist_credit_name acn) {
    int artist_credit_name_id = get_artist_credit_name_id(db, acn.artist_credit, acn.position, acn.artist, acn.name);
    if (artist_credit_name_id != 0) {
        return artist_credit_name_id;
    }
    std::string query = "INSERT INTO artist_credit_name (artist_credit, position, artist, name) VALUES (?, ?, ?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_bind_int(stmt, 1, acn.artist_credit);
    sqlite3_bind_int(stmt, 2, acn.position);
    sqlite3_bind_int(stmt, 3, acn.artist);
    sqlite3_bind_text(stmt, 4, acn.name.c_str(), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    artist_credit_name_id = sqlite3_last_insert_rowid(db);
    return artist_credit_name_id;
}
int get_type_id(sqlite3 *db, std::string type) {
    std::string query = "SELECT id FROM release_group_primary_type WHERE name = '" + type + "'";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_step(stmt);
    int type_id = sqlite3_column_int(stmt, 0);
    return type_id;
}
int add_type(sqlite3 *db, std::string type) {
    int type_id = get_type_id(db, type);
    if (type_id != 0) {
        return type_id;
    }
    std::string query = "INSERT INTO type (name) VALUES (?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, type.c_str(), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    type_id = sqlite3_last_insert_rowid(db);
    return type_id;
}
int get_release_id(sqlite3 *db, std::string release_mbid) {
    std::string query = "SELECT id FROM release WHERE mbid = '" + release_mbid + "'";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_step(stmt);
    int release_id = sqlite3_column_int(stmt, 0);
    return release_id;
}
int add_release(sqlite3 *db, release r) {
    int release_id = get_release_id(db, r.mbid);
    if (release_id != 0) {
        return release_id;
    }
    std::string query = "INSERT INTO release (mbid, name, artist_credit, date, type, cover_url) VALUES (?, ?, ?, ?, ?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, r.mbid.c_str(), -1, SQLITE_STATIC);
    sqlite3_bind_text(stmt, 2, r.name.c_str(), -1, SQLITE_STATIC);
    sqlite3_bind_int(stmt, 3, r.artist_credit);
    sqlite3_bind_int(stmt, 4, r.date);
    sqlite3_bind_int(stmt, 5, r.type);
    sqlite3_bind_text(stmt, 6, r.cover_url.c_str(), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    release_id = sqlite3_last_insert_rowid(db);
    return release_id;
}
int add_track(sqlite3 *db, track t) {
    std::string query = "INSERT INTO track (mbid, name, number, artist_credit, length, release, url) VALUES (?, ?, ?, ?, ?, ?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query.c_str(), -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, t.mbid.c_str(), -1, SQLITE_STATIC);
    sqlite3_bind_text(stmt, 2, t.name.c_str(), -1, SQLITE_STATIC);
    sqlite3_bind_int(stmt, 3, t.number);
    sqlite3_bind_int(stmt, 4, t.artist_credit);
    sqlite3_bind_int(stmt, 5, t.length);
    sqlite3_bind_int(stmt, 6, t.release);
    sqlite3_bind_text(stmt, 7, t.url.c_str(), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    int track_id = sqlite3_last_insert_rowid(db);
    return track_id;
}

void processAudio(std::filesystem::path path, sqlite3 *db)
{
    const char *filename = path.c_str();
    TagLib::FileRef f(filename);
    TagLib::PropertyMap tags = f.file()->properties();
    int artist_count = tags["ARTISTS"].size();
    std::string artist_credit = tags["ARTIST"][0].to8Bit(true);
    int artist_credit_id = add_artist_credit(db, artist_credit, artist_count);
    artist_credit_name artist_credit_names[artist_count];
    for (int i = 0; i < artist_count; i++) {
        artist_credit_names[i].artist = add_artist(db, tags["MUSICBRAINZ_ARTISTID"][i].to8Bit(true));
        artist_credit_names[i].artist_credit = artist_credit_id;
        artist_credit_names[i].position = i;
        artist_credit_names[i].name = tags["ARTISTS"][i].to8Bit(true);
        add_artist_credit_name(db, artist_credit_names[i]);
    }
    int type_id = add_type(db, tags["RELEASETYPE"][0].to8Bit(true));
    int album_artist_count = tags["ALBUMARTISTS"].size();
    int album_artist_credit_id = add_artist_credit(db, tags["ALBUMARTIST"][0].to8Bit(true), album_artist_count);
    artist_credit_name album_artist_credit_names[album_artist_count];
    for (int i = 0; i < album_artist_count; i++) {
        album_artist_credit_names[i].artist = add_artist(db, tags["MUSICBRAINZ_ALBUMARTISTID"][i].to8Bit(true));
        album_artist_credit_names[i].artist_credit = album_artist_credit_id;
        album_artist_credit_names[i].position = i;
        album_artist_credit_names[i].name = tags["ALBUMARTISTS"][i].to8Bit(true);
        add_artist_credit_name(db, album_artist_credit_names[i]);
    }
    release r;
    r.mbid = tags["MUSICBRAINZ_ALBUMID"][0].to8Bit(true);
    r.name = tags["ALBUM"][0].to8Bit(true);
    r.artist_credit = album_artist_credit_id;
    if (tags["DATE"].size() > 0) {
        r.date = atoi(tags["DATE"][0].to8Bit(true).c_str());
    } else {
        r.date = 0;
    }
    r.type = type_id;
    std::string filename_str = filename;
    std::filesystem::path parent_dir = path.parent_path();
    for (const auto & entry : std::filesystem::directory_iterator(parent_dir)) {
        std::string entry_filename = entry.path().filename().string();
        if (entry_filename.find("cover") != std::string::npos) {
            r.cover_url = entry.path().string();
            break;
        }
    }


    int release_id = add_release(db, r);
    
    track t;
    t.mbid = tags["MUSICBRAINZ_TRACKID"][0].to8Bit(true);
    t.name = tags["TITLE"][0].to8Bit(true);
    t.number = std::stoi(tags["TRACKNUMBER"][0].to8Bit(true));
    t.artist_credit = artist_credit_id;
    t.length = f.audioProperties()->lengthInMilliseconds();
    t.release = release_id;
    t.url = filename_str;
    add_track(db, t);
}
int main(int argc, char *argv[]) {
    //Create and open an sqlite database
    sqlite3 *db;
    if (sqlite3_open("../mb.db", &db) == SQLITE_OK) {
    //Setup the database
        std::ifstream createTables("CreateTables.sql");
        std::string sql = std::string(std::istreambuf_iterator<char>(createTables), std::istreambuf_iterator<char>());
        sqlite3_exec(db, sql.c_str(), NULL, NULL, NULL);
    } else {
        printf("Error opening database");
    }
    std::string path = argv[1];
    for (const auto & entry : std::filesystem::recursive_directory_iterator(path)) {
        if (!is_directory(entry)) {
            std::string filename = entry.path().filename();
            if (audio_ext.find(filename.substr(filename.find_last_of(".") + 1)) != std::string::npos) {
                processAudio(entry.path(), db);
            }
        }
    }
    sqlite3_close(db);
}
