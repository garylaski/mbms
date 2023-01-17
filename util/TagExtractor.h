#include "DatabaseManager.h"
#include <taglib/tag.h>
#include <taglib/tpropertymap.h>
#include <string>
class TagExtractor
{
public:
    TagExtractor(DatabaseManager* const);
    ~TagExtractor();
    void extract(char const* filename, char const* cover_url);
private:
    DatabaseManager* m_database_manager;
    TagLib::PropertyMap m_properties;
    Track m_track = Track();
    Artist m_artist = Artist();
    Release m_release = Release();
    Type m_type = Type();
    ArtistCredit m_artist_credit = ArtistCredit();
    ArtistCreditName m_artist_credit_name = ArtistCreditName();
    char const* m_image_extensions[] = {".jpg", ".jpeg", ".png", ".gif"};
    std::filesystem::path m_cover_path;
};
