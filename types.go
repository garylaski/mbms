package main

import "log"

type Format struct {
    FormatName       string      `json:"format_name"`
    Tags *Tags `json:"tags,omitempty"`
}
type Stream struct {
    Duration float64 `json:"duration,string"`
    Tags    *Tags   `json:"tags,omitempty"`
}
type Ffprobe struct {
    Format  *Format   `json:"format"`
    Streams  []*Stream   `json:"streams,omitempty"`
}
type Tags struct {
    Artist            string      `json:"ARTIST"`
    Artists           string      `json:"ARTISTS"`
    Title             string      `json:"TITLE"`
    MusicBrainzTrackID string      `json:"MUSICBRAINZ_RELEASETRACKID,omitempty"`
    MusicBrainzAlbumID string      `json:"MUSICBRAINZ_ALBUMID,omitempty"`
    MusicBrainzArtistID string      `json:"MUSICBRAINZ_ARTISTID,omitempty"`
    MusicBrainzAlbumArtistID string      `json:"MUSICBRAINZ_ALBUMARTISTID,omitempty"`
    AlbumArtist      string      `json:"ALBUM_ARTIST,omitempty"`
    AlbumArtists      string      `json:"ALBUMARTISTS"`
    Track             string      `json:"TRACK"`    
    Album             string      `json:"ALBUM"`
    Date              string      `json:"DATE,omitempty"`
    AlbumArtistSpaced string      `json:"ALBUM ARTIST,omitempty"`
    MusicBrainzTrackIDSpaced string `json:"MusicBrainz Release Track Id,omitempty"`
    MusicBrainzAlbumIDSpaced string      `json:"MusicBrainz Album Id,omitempty"`
    MusicBrainzArtistIDSpaced string      `json:"MusicBrainz Artist Id,omitempty"`
    MusicBrainzAlbumArtistIDSpaced string      `json:"MusicBrainz Album Artist Id,omitempty"`
    OriginalDate      string      `json:"ORIGINALDATE,omitempty"`
    Tdor              string      `json:"TDOR,omitempty"`
    Year              string      `json:"YEAR,omitempty"`
}


func standardizeTags(ffprobe *Ffprobe) {
    if (ffprobe.Format.Tags == nil) {
        if (len(ffprobe.Streams) > 0) {
            ffprobe.Format.Tags = ffprobe.Streams[0].Tags
        } else {
            log.Printf("No tags found in ffprobe output for %s\n", ffprobe.Format.FormatName)
            return
        }
    }
    tags := ffprobe.Format.Tags
    if (tags == nil) {
        log.Printf("No tags found in ffprobe output for %s\n", ffprobe.Format.FormatName)
        return
    }
    if (tags.MusicBrainzTrackID == "") {
        if (tags.MusicBrainzTrackIDSpaced != "") {
            tags.MusicBrainzTrackID = tags.MusicBrainzTrackIDSpaced
        }
    }
    if (tags.MusicBrainzAlbumID == "") {
        if (tags.MusicBrainzAlbumIDSpaced != "") {
            tags.MusicBrainzAlbumID = tags.MusicBrainzAlbumIDSpaced
        }
    }
    if (tags.MusicBrainzArtistID == "") {
        if (tags.MusicBrainzArtistIDSpaced != "") {
            tags.MusicBrainzArtistID = tags.MusicBrainzArtistIDSpaced
        }
    }
    if (tags.MusicBrainzAlbumArtistID == "") {
        if (tags.MusicBrainzAlbumArtistIDSpaced != "") {
            tags.MusicBrainzAlbumArtistID = tags.MusicBrainzAlbumArtistIDSpaced
        }
    }
    if (tags.AlbumArtist == "") {
        if (tags.AlbumArtistSpaced != "") {
            tags.AlbumArtist = tags.AlbumArtistSpaced
        }
    }
    if (tags.Date == "") {
        if (tags.OriginalDate != "") {
            tags.Date = tags.OriginalDate
        } else if (tags.Tdor != "") {
            tags.Date = tags.Tdor
        } else if (tags.Year != "") {
            tags.Date = tags.Year
        }
    }
}
