#include "DatabaseManager.h"
#include "TagExtractor.h"
#include <drogon/drogon.h>

int main() {
    DatabaseManager db("/home/gary/repos/mbms/mb.db");
    TagExtractor tag(&db);
    //tag.extract_path("/backup/source");
    //db.update_artist_info();
    drogon::app().addListener("0.0.0.0",8080);
    drogon::app().run();
    return 0;
}
