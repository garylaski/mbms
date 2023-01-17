#include <taglib/fileref.h>
#include <taglib/tag.h>
#include <taglib/tpropertymap.h>
#include <taglib/tstring.h>
#include <taglib/tstringlist.h>
#include <taglib/tfile.h>
#include <curl/curl.h>
#include "TagExtractor.h"
#include <filesystem>

TagExtractor::TagExtractor(DatabaseManager* database_manager):
    m_database_manager(database_manager) {}

    TagExtractor::~TagExtractor() {}

    void TagExtractor::extract(std::filesystem::path* path) {
        TagLib::FileRef f(path->string().c_str());
        m_track.length = f.audioProperties()->lengthInMilliseconds();
        m_properties = f.file()->properties();
        m_release.mbid = m_properties["MUSICBRAINZ_ALBUMID"][0].toCString();
        if ((m_track.release = m_database_manager->get_release_id(m_release.mbid)) == 0) {
            m_artist_credit.name = m_properties["ALBUMARTIST"][0].toCString(true);
            if ((m_release.artist_credit = m_database_manager->get_artist_credit_id(m_artist_credit.name)) == 0) {
                m_release.artist_credit = m_database_manager->add_artist_credit(&m_artist_credit);
                for (int i = 0; i < m_properties["MUSICBRAINZ_ALBUMARTISTID"].size(); i++) {
                    m_artist.mbid = m_properties["MUSICBRAINZ_ALBUMARTISTID"][i].toCString();
                    m_artist_credit_name.artist_credit = m_release.artist_credit;
                    m_artist_credit_name.name = m_properties["ALBUMARTISTS"][i].toCString(true);
                    if ((m_artist_credit_name.artist = m_database_manager->get_artist_id(m_artist.mbid)) == 0) {
                        m_artist_credit_name.artist = m_database_manager->add_artist(&m_artist);
                    }
                    m_database_manager->add_artist_credit_name(&m_artist_credit_name);
                }
            }
            m_type.name = m_properties["MUSICBRAINZ_ALBUMTYPE"].toString().toCString(true);
            m_release.type = m_database_manager->add_release_type(&m_type);
            m_release.name = m_properties["ALBUM"][0].toCString(true);
            m_release.cover_url = find_cover_url(path);
            m_release.date = m_properties["DATE"][0].toCString();
            m_track.release = m_database_manager->add_release(&m_release);
        }
        m_track.mbid = m_properties["MUSICBRAINZ_TRACKID"][0].toCString();
        m_track.name = m_properties["TITLE"][0].toCString(true);
        m_track.url = curl_escape(filename, 0);
        m_track.number = m_properties["TRACKNUMBER"][0].toInt();
        m_artist_credit.name = m_properties["ARTIST"][0].toCString(true);
        if ((m_track.artist_credit = m_database_manager->get_artist_credit_id(m_artist_credit.name)) == 0) {
            m_track.artist_credit = m_database_manager->add_artist_credit(&m_artist_credit);
            for (int i = 0; i < m_properties["MUSICBRAINZ_ARTISTID"].size(); i++) {
                m_artist.mbid = m_properties["MUSICBRAINZ_ARTISTID"][i].toCString();
                m_artist_credit_name.artist_credit = m_track.artist_credit;
                m_artist_credit_name.name = m_properties["ARTISTS"][i].toCString(true);
                if ((m_artist_credit_name.artist = m_database_manager->get_artist_id(m_artist.mbid)) == 0) {
                    m_artist_credit_name.artist = m_database_manager->add_artist(&m_artist);
                }
                m_database_manager->add_artist_credit_name(&m_artist_credit_name);
            }
        }
        m_database_manager->add_track(&m_track);
    }
char const* TagExtractor::find_cover_url(std::filesystem::path* path) {
    char const* image_extensions[] = {".jpg", ".jpeg", ".png", ".gif"};
    for (auto extension : m_image_extensions) {
        m_cover_path = path / "cover" / extension;
        if (std::filesystem::exists(m_cover_path))
            return cover_path.string().c_str();
    }
    return "";
}
