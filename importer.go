package main

import (
    "log"
    "path/filepath"
    "os/exec"
    "os"
    "sync"
    "strings"
    "strconv" 
    "encoding/json"
    "bytes"
)

type Release struct {
    mbid string
    name string
    artist_credit int
    date string
    cover string
    wg sync.WaitGroup
    set bool
}

func (server *Server) ImportLibrary(path string) {
    defer server.wg.Done()
    dirs, err := os.ReadDir(path)
    if err != nil {
        log.Printf("ImportLibrary: %v\n", err)
    }
    var release *Release
    for _, dir := range dirs {
        if dir.IsDir() {
            server.wg.Add(1)
            go server.ImportLibrary(filepath.Join(path, dir.Name()))
        } else {
            if release == nil {
                release = &Release{
                    wg: sync.WaitGroup{},
                    set: false,
                }
                server.wg.Add(1)
                go server.AddRelease(release)
            }
            release.wg.Add(1)
            go server.ProcessFile(filepath.Join(path, dir.Name()), release)
        }
    }
}

func (server *Server) ProcessFile(path string, release *Release) {
    cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json=compact=1", "-show_entries", "format=format_name:stream=duration:stream_tags:format_tags", path)
    out := server.bufferPool.Get().(*bytes.Buffer)
    out.Reset()
    cmd.Stdout = out
    err := cmd.Run()
    if err != nil {
        log.Printf("ImportFileToDB: ffprobe %v : %v:\n%v\n", path, err, out)
    }
    data := server.ffprobePool.Get().(Ffprobe)
    defer server.ffprobePool.Put(data)
    err = json.Unmarshal(out.Bytes(), &data)
    server.bufferPool.Put(out)
    if err != nil {
        log.Printf("ImportFileToDB: %v\n", err)
    }
    switch data.Format.FormatName {
    case "ogg", "flac", "aac":
        standardizeTags(&data)
        server.AddTrack(path, &data, release)
    case "image2", "jpeg_pipe", "png_pipe":
        release.cover = path
        release.wg.Done()
    default:
        release.wg.Done()
    }
}

func (server *Server) AddTrack(path string, data *Ffprobe, release *Release) {
    if !release.set {
        release.artist_credit = server.AddArtistCredit(&data.Format.Tags.AlbumArtist)
        server.AddArtistCreditNames(release.artist_credit, &data.Format.Tags.MusicBrainzAlbumArtistID, &data.Format.Tags.AlbumArtists)
        release.name = data.Format.Tags.Album
        release.date = data.Format.Tags.Date
        release.mbid = data.Format.Tags.MusicBrainzAlbumID
        release.set = true
    }
    release.wg.Done()
    idx := strings.Index(data.Format.Tags.Track, "/")
    if idx == -1 {
        idx = len(data.Format.Tags.Track)
    }
    number, _ := strconv.Atoi(data.Format.Tags.Track[:idx])
    artist_credit := server.AddArtistCredit(&data.Format.Tags.Artist)
    if (data.Format.Tags.MusicBrainzArtistID == "") {
        log.Fatal(path)
    }
    server.AddArtistCreditNames(artist_credit, &data.Format.Tags.MusicBrainzArtistID, &data.Format.Tags.Artists)
    server.batch.Queue("INSERT INTO track (name, mbid, number, length, release_mbid, url, artist_credit) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (mbid) DO NOTHING",
    data.Format.Tags.Title, data.Format.Tags.MusicBrainzTrackID, number, int(data.Streams[0].Duration * 1000), data.Format.Tags.MusicBrainzAlbumID, path, artist_credit)
}


func (server *Server) AddArtistCredit(name *string) int {
    if id, ok := server.artistCreditCache.Load(*name); ok {
        return id.(int)
    }
    id := server.artistCreditId
    server.artistCreditId++
    server.batch.Queue("INSERT INTO artist_credit (name, id) VALUES ($1, $2)",
    name, id)
    server.artistCreditCache.Store(*name, id)
    return id
}

func (server *Server) AddArtistCreditNames(artist_credit int, mbids *string, names *string) {
    mbid := strings.Split(*mbids, ";")
    name := strings.Split(*names, ";")
    if len(mbid) != len(name) {
        log.Printf("%v, %v\n", mbid, name)
        log.Printf("%v, %v\n", *mbids, *names)
        log.Println("mbid and name length mismatch")
        return
    }
    for i := 0; i < len(mbid); i++ {
        if _, ok := server.artistCreditNameCache.LoadOrStore(Ac{mbid[i], artist_credit}, true); ok {
            continue
        }
        server.batch.Queue("INSERT INTO artist_credit_name (artist_credit, artist_mbid, name) VALUES ($1, $2, $3) ON CONFLICT (artist_credit, artist_mbid) DO NOTHING",
        artist_credit, mbid[i], name[i])
        if _, ok := server.artistCache.LoadOrStore(mbid[i], true); ok {
            continue
        }
        server.batch.Queue("INSERT INTO artist (mbid, name) VALUES ($1, $2)",
        mbid[i], name[i])
    }
}

func (server *Server) AddRelease(release *Release) {
    defer server.wg.Done()
    release.wg.Wait()
    server.batch.Queue("INSERT INTO release (mbid, name, artist_credit, date, cover_url) VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING",
    release.mbid, release.name, release.artist_credit, release.date, release.cover)
}

