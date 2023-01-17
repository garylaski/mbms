#pragma once
struct Track {
    char const* mbid { nullptr };
    char const* name { nullptr };
    char const* url { nullptr };
    int number { 0 };
    int artist_credit { 0 };
    int length { 0 };
    int release { 0 };
};
struct Artist {
    char const* mbid { nullptr };
    char const* name { nullptr };
};
struct Release {
    char const* mbid { nullptr };
    char const* name { nullptr };
    char const* date { nullptr };
    char const* cover_url { nullptr };
    int artist_credit { 0 };
    int type { 0 };
};
struct Type {
    char const* name { nullptr };
};
struct ArtistCredit {
    char const* name { nullptr };
};
struct ArtistCreditName {
    int artist_credit { 0 };
    int artist { 0 };
    char const* name { nullptr };
};
