#include <drogon/HttpController.h>
using namespace drogon;
class DefaultController : public drogon::HttpController<DefaultController>
{
    public:
        METHOD_LIST_BEGIN
        ADD_METHOD_TO(DefaultController::getRelease, "/release/{1}", Get);
        ADD_METHOD_TO(DefaultController::getArtist, "/artist/{1}", Get);
        ADD_METHOD_TO(DefaultController::getTrack, "/track/{1}", Get);
        METHOD_LIST_END
        void getRelease(const HttpRequestPtr &req, 
                std::function<void(const HttpResponsePtr &)> &&callback, 
                std::string mbid) const;
        void getArtist(const HttpRequestPtr &req, 
                std::function<void(const HttpResponsePtr &)> &&callback, 
                std::string mbid) const;
        void getTrack(const HttpRequestPtr &req, 
                std::function<void(const HttpResponsePtr &)> &&callback, 
                std::string mbid) const;
    private:
        DatabaseManager* m_database_manager;
};
