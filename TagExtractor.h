#include "DatabaseManager.h"
#include <taglib/tag.h>
#include <taglib/tpropertymap.h>
#include <string>
#include <filesystem>
#include <string>
class TagExtractor
{
public:
    TagExtractor(DatabaseManager* const);
    ~TagExtractor();
    void extract(std::filesystem::path const*);
    void extract_path(char const*);
private:
    DatabaseManager* m_database_manager;
    TagLib::PropertyMap m_properties;
    Track m_track = Track();
    Artist m_artist = Artist();
    Release m_release = Release();
    Type m_type = Type();
    ArtistCredit m_artist_credit = ArtistCredit();
    ArtistCreditName m_artist_credit_name = ArtistCreditName();
    char const* m_cover_name = "/cover.";
    std::string m_cover_path;
    char const* m_filename;
    constexpr static char const* m_image_extensions[4] = {"jpg", "jpeg", "png", "gif"};
    char const* find_cover_url(char const*);
};
