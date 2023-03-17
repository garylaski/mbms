#include "DatabaseManager.h"
#include "TagExtractor.h"
#include "oatpp/network/Server.hpp"

int main() {
    DatabaseManager db("/home/gary/repos/mbms/mb.db");
    /*
    TagExtractor tag(&db);
    tag.extract_path("/backup/source");
    db.update_artist_info();
    */
    oatpp::base::Environment::init();
    oatpp::network::Server server(connectionProvider, connectionHandler);



}
