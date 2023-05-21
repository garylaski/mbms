#include "DefaultController.h"
#include "DatabaseManager.h"

void DefaultController::getRelease(const HttpRequestPtr &req, std::function<void(const HttpResponsePtr &)> &&callback, std::string mbid) const {
    auto resp = HttpResponse::newHttpResponse();
    std::string name;
    
    callback(resp);
}
