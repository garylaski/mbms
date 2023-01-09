1. backend with sql database that stores all info.
    - Generate a partial MB DB
    - Import media files into DB, Using default tags that Picard writes
        -There seems to be an issue here with Album Artists Multitag (picard does not write by default).
        -Could there be another way? Using mb album artist ids?
    - Add files to PostgreSQL DB, usetaglib to get tags
    - (additional tags for data urls)

2. server that generates pages from database
    - retains all info given by MBDB
    - Artist "canon" name needs to be determined. Either by MB API query or inference on DB.
