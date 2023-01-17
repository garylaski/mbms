#include "DatabaseManager.h"
#include "TagExtractor.h"
#include <iostream>
#include <filesystem>
bool is_audio_file(const std::filesystem::path ext)
{
    return ext == ".mp3" || ext == ".wav" || ext == ".ogg" || ext == ".flac";
}
bool is_image_file(const std::filesystem::path ext)
{
    return ext == ".jpg" || ext == ".png";
}
int main(int argc, char** argv)
{
    if (argc != 2)
    {
        std::cout << "Usage: " << argv[0] << " <music directory>" << std::endl;
        return 1;
    }
    DatabaseManager database_manager("mbms.sqlite");
    TagExtractor tag_extractor(&database_manager);
    for (const auto & entry : std::filesystem::recursive_directory_iterator(argv[1]))
    {
        if (entry.is_regular_file())
        {
            if (is_audio_file(entry.path().extension()))
            {
                tag_extractor.extract(entry.path());
            } 
        }
    }
}
