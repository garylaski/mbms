#include <taglib/tag.h>
#include <taglib/fileref.h>
#include <taglib/tfilestream.h>
#include <taglib/tpropertymap.h>
#include <sqlite3.h>
#include <stdio.h>
#include <fstream>
#include <string>
#include <filesystem>
#include <map>
#include <string>
#include <iostream>
#include <curl/curl.h>
#include <libxml/parser.h>
#include <libxml/tree.h>
#include <libxml/xpath.h>
#include <libxml/xpathInternals.h>
#include <cstring>
#include <chrono>
#include <thread>
using namespace std;
struct {
    map<TagLib::String, int> artists;
    map<TagLib::String, int> releases;
    map<TagLib::String, int> artist_credits;
    map<TagLib::String, int> types;
    map<TagLib::String, int> artist_artist_relations;
    map<TagLib::String, int> artist_artist_relations_types;
} cache;
string audio_ext = "flacaacmp3opusm4a";
sqlite3 *db;
CURL *curl = curl_easy_init();
CURLcode res;
struct curl_slist *headers = NULL;
char* date = "0000-00-00";
size_t processArtist(const char* contents, size_t size, size_t nmemb, string *mbid) {
    size_t realsize = size * nmemb;
    string name;
    // xml parsing
    xmlDoc* doc = xmlReadMemory(contents, realsize, "noname.xml", NULL, 0);
    xmlXPathContext* xpathCtx = xmlXPathNewContext(doc);
    xmlXPathRegisterNs(xpathCtx, (const xmlChar*)"mb", (const xmlChar*)"http://musicbrainz.org/ns/mmd-2.0#");
    xmlXPathObject* xpathObj = xmlXPathEvalExpression((xmlChar*) "/mb:metadata/mb:artist/mb:name", xpathCtx);
    xmlNodeSet* nodes = xpathObj->nodesetval;
    for (int i = 0; i < nodes->nodeNr; i++) {
        xmlNode* node = nodes->nodeTab[i];
        if (node->type == XML_ELEMENT_NODE) {
            *name = (char*) node->children->content;
        }
    }
    xmlXPathFreeObject(xpathObj);
    xmlXPathFreeContext(xpathCtx);
    xmlFreeDoc(doc);
    xmlCleanupParser();
    // copy to name
    return realsize;
}
int addArtistArtistRelationType(string type) {
    const char *query = "INSERT INTO artist_artist_relation_types (name) VALUES (?);";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, type.c_str(), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    return sqlite3_last_insert_rowid(db);
}
int addArtistArtistRelation(int artist_id, int related_artist_id, string artist_artist_relation_type) {
    int artist_artist_relation_type_id;
    if ((artist_artist_relation_type_id = cache.artist_artist_relations_types[artist_artist_relation_type]) == 0) {
        artist_artist_relation_type_id = addArtistArtistRelationType(artist_artist_relation_type);
        cache.artist_artist_relations_types[artist_artist_relation_type] = artist_artist_relation_type_id;
    }
    const char* query = "INSERT INTO artist_artist_rel (artist, related_artist, artist_artist_relation_type) VALUES (?, ?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_int(stmt, 1, artist_id);
    sqlite3_bind_int(stmt, 2, related_artist_id);
    sqlite3_bind_int(stmt, 3, artist_artist_relation_type_id);
    sqlite3_step(stmt);
    return sqlite3_last_insert_rowid(db);
}
int addTrack(TagLib::PropertyMap tags, int artist_credit_id, int length, int release_id, const char* url) {
    const char* query = "INSERT INTO track (mbid, name, number, artist_credit, length, release, url) VALUES (?, ?, ?, ?, ?, ?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, tags["MUSICBRAINZ_TRACKID"][0].toCString(), -1, SQLITE_STATIC);
    sqlite3_bind_text(stmt, 2, tags["TITLE"][0].toCString(true), -1, SQLITE_STATIC);
    sqlite3_bind_int(stmt, 3, tags["TRACKNUMBER"][0].toInt());
    sqlite3_bind_int(stmt, 4, artist_credit_id);
    sqlite3_bind_int(stmt, 5, length);
    sqlite3_bind_int(stmt, 6, release_id);
    sqlite3_bind_text(stmt, 7, url, -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    return sqlite3_last_insert_rowid(db);
}
int addArtist(TagLib::String mbid) {
    const char* query = "INSERT INTO artist (mbid) VALUES (?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, mbid.toCString(), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    return sqlite3_last_insert_rowid(db);
}
int addArtistName(string mbid, string name) {
    cout << "Adding artist name: " << name << endl;
    const char* query = "UPDATE artist SET name = ? WHERE mbid = ?";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, name.c_str(), -1, SQLITE_STATIC);
    sqlite3_bind_text(stmt, 2, mbid.c_str(), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    return sqlite3_last_insert_rowid(db);
}
void queryMB_Artist(string mbid, int artist_id) {
    string url = "http://musicbrainz.org/ws/2/artist/" + mbid + "?inc=artist-rels";
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
        curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, processArtist);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &mbid);
        res = curl_easy_perform(curl);
        if (res != CURLE_OK) {
            fprintf(stderr, "curl_easy_perform() failed: %s\n", curl_easy_strerror(res));
        }
    }
}
void addArtistCreditName(int artist_credit_id, int position, int artist_id, TagLib::String name) {
    const char* query = "INSERT INTO artist_credit_name (artist_credit, position, artist, name) VALUES (?, ?, ?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_int(stmt, 1, artist_credit_id);
    sqlite3_bind_int(stmt, 2, position);
    sqlite3_bind_int(stmt, 3, artist_id);
    sqlite3_bind_text(stmt, 4, name.toCString(true), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
}
int addArtistCredit(TagLib::String artist_credit, TagLib::StringList artist_credit_names, TagLib::StringList artist_mbids) {
    int artist_count = artist_credit_names.size();
    const char* query = "INSERT INTO artist_credit (name, artist_count) VALUES (?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, artist_credit.toCString(true), -1, SQLITE_STATIC);
    sqlite3_bind_int(stmt, 2, artist_count);
    sqlite3_step(stmt);
    int artist_credit_id = sqlite3_last_insert_rowid(db);
    for (int i = 0; i < artist_count; i++) {
        int artist_id;
        TagLib::String artist_mbid = artist_mbids[i];
        if ((artist_id = cache.artists[artist_mbid]) == 0) {
            artist_id = addArtist(artist_mbid);
            cache.artists[artist_mbid] = artist_id;
        }
        addArtistCreditName(artist_credit_id, i, artist_id, artist_credit_names[i]);
    }
    return artist_credit_id;
}
int addType(TagLib::String type) {
    const char* query = "INSERT INTO type (name) VALUES (?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, type.toCString(true), -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    return sqlite3_last_insert_rowid(db);
}
int addRelease(TagLib::PropertyMap tags, TagLib::String release_mbid, const char* cover_url) {
    int album_artist_credit_id, type_id;
    TagLib::String album_artist_credit = tags["ALBUMARTIST"][0];
    if ((album_artist_credit_id = cache.artist_credits[album_artist_credit]) == 0){
        album_artist_credit_id = addArtistCredit(album_artist_credit, tags["ALBUMARTISTS"], tags["MUSICBRAINZ_ALBUMARTISTID"]);
        cache.artist_credits[album_artist_credit] = album_artist_credit_id;
    }
    TagLib::String type = tags["RELEASETYPE"][0];
    if ((type_id = cache.types[type]) == 0) {
        type_id = addType(type);
        cache.types[type] = type_id;
    }
    const char* query = "INSERT INTO release (mbid, name, artist_credit, date, type, cover_url) VALUES (?, ?, ?, ?, ?, ?)";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    sqlite3_bind_text(stmt, 1, release_mbid.toCString(), -1, SQLITE_STATIC);
    sqlite3_bind_text(stmt, 2, tags["ALBUM"][0].toCString(true), -1, SQLITE_STATIC);
    sqlite3_bind_int(stmt, 3, album_artist_credit_id);
    if (tags["DATE"].size() > 0) {
        sqlite3_bind_text(stmt, 4, tags["DATE"][0].toCString(true), -1, SQLITE_STATIC);
    } else  {
        sqlite3_bind_text(stmt, 4, date, -1, SQLITE_STATIC);
    }
    sqlite3_bind_int(stmt, 5, type_id);
    sqlite3_bind_text(stmt, 6, cover_url, -1, SQLITE_STATIC);
    sqlite3_step(stmt);
    return sqlite3_last_insert_rowid(db);
}
void processFiles(string path) {
    for (const auto & entry : filesystem::recursive_directory_iterator(path)) {
        if (!is_directory(entry)) {
            string filename = entry.path().string();
            if (audio_ext.find(filename.substr(filename.find_last_of(".") + 1)) != string::npos) {
                TagLib::FileStream stream(filename.c_str(), true);
                TagLib::FileRef f(&stream);
                TagLib::PropertyMap tags = f.file()->properties();
                int length = f.audioProperties()->lengthInMilliseconds();
                int release_id, artist_credit_id;
                TagLib::String release_mbid = tags["MUSICBRAINZ_ALBUMID"][0];
                if ((release_id = cache.releases[release_mbid]) == 0) {
                    string cover_url = entry.path().parent_path().string() + "/cover";
                    // iterate over common image extensions
                    if (filesystem::exists(cover_url + ".jpg")) {
                        cover_url = cover_url + ".jpg";
                    } else if (filesystem::exists(cover_url + ".jpeg")) {
                        cover_url = cover_url + ".jpeg";
                    } else if (filesystem::exists(cover_url + ".png")) {
                        cover_url = cover_url + ".png";
                    } else if (filesystem::exists(cover_url + ".gif")) {
                        cover_url = cover_url + ".gif";
                    } else if (filesystem::exists(cover_url + ".bmp")) {
                        cover_url = cover_url + ".bmp";
                    }
                    const char* enc_url = curl_easy_escape(curl, cover_url.c_str(), cover_url.length());
                    release_id = addRelease(tags, release_mbid, enc_url);
                    cache.releases[release_mbid] = release_id;
                }
                TagLib::String artist_credit = tags["ARTIST"][0];
                if ((artist_credit_id = cache.artist_credits[artist_credit]) == 0) {
                    artist_credit_id = addArtistCredit(artist_credit, tags["ARTISTS"], tags["MUSICBRAINZ_ARTISTID"]);
                    cache.artist_credits[artist_credit] = artist_credit_id;
                }
                const char* enc_url = curl_easy_escape(curl, filename.c_str(), filename.length());
                cout << "Adding " << filename << endl;
                addTrack(tags, artist_credit_id, length, release_id, enc_url);
            }
        }
    }
}

void postProcess() {
    if (curl) {
        headers = curl_slist_append(headers, "User-Agent: mbms/0.0.1");
    }
    string artist_mbid;
    int artist_id;
    const char* query = "SELECT id, mbid FROM artist";
    sqlite3_stmt *stmt;
    sqlite3_prepare_v2(db, query, -1, &stmt, NULL);
    // iterate over all artists and add their name
    while (sqlite3_step(stmt) == SQLITE_ROW) {
        int artist_id = sqlite3_column_int(stmt, 0);
        artist_mbid = string(reinterpret_cast<const char*> (sqlite3_column_text(stmt, 1)));
        cout << "Processing artist " << artist_mbid << endl;
        queryMB_Artist(artist_mbid, artist_id);
        this_thread::sleep_for(chrono::milliseconds(1000));
        this_thread::sleep_until(chrono::system_clock::now() + chrono::milliseconds(1000));
    }
    sqlite3_finalize(stmt);
}
int main(int argc, char *argv[]) {
    sqlite3 *file;
    sqlite3_open("../mb.db", &file);
    if (sqlite3_open(":memory:", &db) == SQLITE_OK) {
        ifstream createTables("CreateTables.sql");
        string sql = string(istreambuf_iterator<char>(createTables), istreambuf_iterator<char>());
        sqlite3_exec(db, sql.c_str(), NULL, NULL, NULL);
    }
    string path = "/backup/source/lossless/Tek Lintowe";
    processFiles(path);
    postProcess();
    sqlite3_backup *pBackup = sqlite3_backup_init(file, "main", db, "main");
    if(pBackup){
      sqlite3_backup_step(pBackup, -1);
      sqlite3_backup_finish(pBackup);
    }
    //save sqlite db to file
    sqlite3_close(db);
    curl_easy_cleanup(curl);
}
