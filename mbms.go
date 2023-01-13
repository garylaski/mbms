package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
    "io"
    "log"
    "net/http"
    "net/url"
    "os"
    "strconv"
    "strings"
)

type Artist struct {
    id       int
    mbid     string
    releases []Release
    appearances []Release
    name     string
}
type ArtistCredit struct {
    id                int
    name              string
    artistCount       int
    artistCreditNames []ArtistCreditName
}
type ArtistCreditName struct {
    artistCreditId int
    position       int
    name           string
    artistId       int
    artistMbid     string
}
type Release struct {
    id             int
    mbid           string
    name           string
    artistCreditId int
    date           string
    _type          int
    coverUrl       string
    tracks         []Track
    artistCredit   ArtistCredit
}
type Track struct {
    id             int
    mbid           string
    name           string
    number         int
    artistCreditId int
    artistCredit   ArtistCredit
    length         int
    release        int
    url            string
    releaseCoverUrl string
}
type Server struct {
    db               *sql.DB
    navHTML          []byte
    playerHTML       []byte
    releaseHTML      []byte
    trackHTML        []byte
    artistHTML       []byte
    smallReleaseHTML []byte
    smallTrackHTML   []byte
    smallArtistHTML  []byte
    allReleaseHTML   []byte
    allTrackHTML     []byte
    allArtistHTML    []byte
    searchHTML       []byte
    headHTML       []byte
}

func (server Server) getNamesFromMBID(mbid string) string {
    url := "https://musicbrainz.org/ws/2/artist/" + mbid + "?fmt=json"
    resp, err := http.Get(url)
    if err != nil {
        log.Fatal(err)
    }
    // get name from json
    var res map[string]interface{}
    restr, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Fatal(err)
    }
    json.Unmarshal([]byte(restr), &res)
    name := res["name"]
    // store in sqlite
    server.db.Exec("UPDATE artist SET name = ? WHERE mbid = ?", name, mbid)
    //convert name to string
    return name.(string)
}
func convertMsToTime(ms int) string {
    sec := ms / 1000
    min := sec / 60
    sec = sec % 60
    return fmt.Sprintf("%d:%02d", min, sec)
}
func (server Server) generateAllTrackHTML() string {
    html := string(server.allTrackHTML)
    html = strings.Replace(html, "{{nav}}", string(server.navHTML), -1)
    html = strings.Replace(html, "{{player}}", string(server.playerHTML), -1)
    // get 10 random tracks from db
    // generate release HTML for each and add to html
    return html
}
func (server Server) generateAllReleaseHTML() string {
    html := string(server.allReleaseHTML)
    html = strings.Replace(html, "{{nav}}", string(server.navHTML), -1)
    html = strings.Replace(html, "{{player}}", string(server.playerHTML), -1)
    // get 10 random releases from db
    // generate release HTML for each and add to html
    return html
}
func (server Server) generateAllArtistHTML() string {
    html := string(server.allArtistHTML)
    html = strings.Replace(html, "{{nav}}", string(server.navHTML), -1)
    html = strings.Replace(html, "{{player}}", string(server.playerHTML), -1)
    // get 10 random artists from db
    // generate release HTML for each and add to html
    return html
}
func (server Server) generateReleaseHTML(release Release) string {
    html := "<html>"
    html += string(server.headHTML)
    html += "<body>"
    html += string(server.navHTML)
    html += string(server.releaseHTML)
    html += string(server.playerHTML)
    html += "</body>"
    html += "</html>"
    html = strings.Replace(html, "{{title}}", release.name, -1)
    html = strings.Replace(html, "{{release.name}}", release.name, -1)
    html = strings.Replace(html, "{{release.date}}", release.date, -1)
    html = strings.Replace(html, "{{release.coverUrl}}", url.PathEscape(release.coverUrl), -1)
    html = strings.Replace(html, "{{release.artistCredit}}", server.generateArtistCreditHTML(release.artistCredit), -1)
    trackHTML := ""
    trackMbids := "["
    for _, track := range release.tracks {
        trackHTML += server.generateSmallTrackHTML(track)
        trackMbids += "'" + track.mbid + "',"
    }
    trackMbids = trackMbids[:len(trackMbids)-1] + "]"
    html = strings.Replace(html, "{{release.tracks}}", trackHTML, -1)
    html = strings.Replace(html, "{{release.tracks.mbid}}", trackMbids, -1)
    return html
}
func (server Server) generateSmallReleaseHTML(release Release) string {
    html := string(server.smallReleaseHTML)
    html = strings.Replace(html, "{{release.name}}", release.name, -1)
    html = strings.Replace(html, "{{release.mbid}}", release.mbid, -1)
    html = strings.Replace(html, "{{release.date}}", release.date, -1)
    html = strings.Replace(html, "{{release.coverUrl}}", url.PathEscape(release.coverUrl), -1)
    html = strings.Replace(html, "{{release.artistCredit}}", server.generateArtistCreditHTML(release.artistCredit), -1)
    return html
}
func (server Server) generateArtistHTML(artist Artist) string {
    html := "<html>"
    html += string(server.headHTML)
    html += "<body>"
    html += string(server.navHTML)
    html += string(server.artistHTML)
    html += string(server.playerHTML)
    html += "</body>"
    html += "</html>"
    if artist.name == "" {
        artist.name = server.getNamesFromMBID(artist.mbid)
    }
    html = strings.Replace(html, "{{title}}", artist.name, -1)
    html = strings.Replace(html, "{{artist.name}}", artist.name, -1)
    var releasesHTML = ""
    for _, release := range artist.releases {
        releasesHTML += server.generateSmallReleaseHTML(release)
    }
    html = strings.Replace(html, "{{artist.releases}}", releasesHTML, -1)
    var appearancesHTML = ""
    for _, appearance := range artist.appearances {
        appearancesHTML += server.generateSmallReleaseHTML(appearance)
    }
    html = strings.Replace(html, "{{artist.appearances}}", appearancesHTML, -1)
    return html
}
func (server Server) generateTrackHTML(track Track) string {
    html := string(server.trackHTML)
    return html
}
func (server Server) generateArtistCreditNameHTML (artistCreditName ArtistCreditName) string {
    html := "<li><a href='/artist/" + artistCreditName.artistMbid + "' onclick=\"ajax('/artist/" + artistCreditName.artistMbid + "'); return false;\">" + artistCreditName.name + "</a></li>"
    return html
}
func (server Server) generateArtistCreditHTML(artistCredit ArtistCredit) string {
    artistCreditNameHTML := artistCredit.name
    for _, artistCreditName := range artistCredit.artistCreditNames {
        artistCreditNameHTML = strings.Replace(artistCreditNameHTML, artistCreditName.name, fmt.Sprintf("<a href=\"/artist/%s\" onclick=\"ajax('/artist/%s'); return false;\">%s</a>", artistCreditName.artistMbid, artistCreditName.artistMbid, artistCreditName.name), -1)
    }
    return artistCreditNameHTML
}
func (server Server) generateSearchHTML(releases []Release, tracks []Track, artistCreditNames []ArtistCreditName) string {
    html := string(server.searchHTML)
    html = strings.Replace(html, "{{nav}}", string(server.navHTML), -1)
    html = strings.Replace(html, "{{player}}", string(server.playerHTML), -1)
    var releasesHTML = ""
    for _, release := range releases {
        releasesHTML += server.generateSmallReleaseHTML(release)
    }
    html = strings.Replace(html, "{{releases}}", releasesHTML, -1)
    var tracksHTML = ""
    for _, track := range tracks {
        tracksHTML += server.generateSmallTrackHTML(track)
    }
    html = strings.Replace(html, "{{tracks}}", tracksHTML, -1)
    var artistCreditNamesHTML = ""
    for _, artistCreditName := range artistCreditNames {
        artistCreditNamesHTML += server.generateArtistCreditNameHTML(artistCreditName)
    }
    html = strings.Replace(html, "{{artistCreditNames}}", artistCreditNamesHTML, -1)
    return html
}
func (server Server) generateSmallTrackHTML(track Track) string {
    html := string(server.smallTrackHTML)
    html = strings.Replace(html, "{{track.name}}", track.name, -1)
    html = strings.Replace(html, "{{track.mbid}}", track.mbid, -1)
    html = strings.Replace(html, "{{track.number}}", strconv.Itoa(track.number), -1)
    html = strings.Replace(html, "{{track.length}}", convertMsToTime(track.length), -1)
    html = strings.Replace(html, "{{track.url}}", url.PathEscape(track.url), -1)
    artistCreditHTML := server.generateArtistCreditHTML(track.artistCredit)
    html = strings.Replace(html, "{{track.artistCredit}}", artistCreditHTML, -1)
    return html
}
func (server Server) artistCreditFromId(id int) ArtistCredit {
    artistCredit := ArtistCredit{}
    query := "SELECT * FROM artist_credit WHERE id = ?"
    row := server.db.QueryRow(query, id)
    row.Scan(&artistCredit.id, &artistCredit.name, &artistCredit.artistCount)
    query = "SELECT * FROM artist_credit_name WHERE artist_credit = ?"
    rows, err := server.db.Query(query, id)
    if err != nil {
        log.Fatal(err)
    }
    for rows.Next() {
        artistCreditName := ArtistCreditName{}
        if err := rows.Scan(&artistCreditName.artistCreditId, &artistCreditName.position, &artistCreditName.artistId, &artistCreditName.name); err != nil {
            log.Printf("Error scanning artist_credit_name")
        }
        artistCredit.artistCreditNames = append(artistCredit.artistCreditNames, artistCreditName)
    }
    query = "SELECT mbid FROM artist WHERE id = ?"
    for i, artistCreditName := range artistCredit.artistCreditNames {
        row := server.db.QueryRow(query, artistCreditName.artistId)
        if err := row.Scan(&artistCredit.artistCreditNames[i].artistMbid); err != nil {
            log.Printf("No artist with id %d", artistCreditName.artistId)
        }
    }
    return artistCredit
}
func artistHandler(w http.ResponseWriter, r *http.Request, server Server) {
    mbid := strings.TrimPrefix(r.URL.Path, "/artist/")
    if mbid == "all" {
        html := server.generateAllArtistHTML()
        fmt.Fprint(w, html)
    } else {
        // query artist with mbid
        artist := Artist{}
        query := "SELECT * FROM artist WHERE mbid = ?"
        row := server.db.QueryRow(query, mbid)
        if err := row.Scan(&artist.id, &artist.mbid, &artist.name); err != nil {
            log.Printf("No artist with mbid %s", mbid)
        }
        query = "SELECT * FROM release WHERE release.artist_credit IN (SELECT artist_credit FROM artist_credit_name WHERE artist = ?) ORDER BY release.date DESC"
        rows, err := server.db.Query(query, artist.id)
        if err != nil {
            log.Fatal(err)
        }
        for rows.Next() {
            var release Release
            err = rows.Scan(&release.id, &release.mbid, &release.name, &release.artistCreditId, &release.date, &release._type, &release.coverUrl)
            if err != nil {
                log.Fatal(err)
            }
            release.artistCredit = server.artistCreditFromId(release.artistCreditId)
            artist.releases = append(artist.releases, release)
        }
        query = "SELECT * FROM release WHERE id IN (SELECT release FROM track WHERE artist_credit IN (SELECT artist_credit FROM artist_credit_name WHERE artist = ?) GROUP BY release) AND release.artist_credit NOT IN (SELECT artist_credit FROM artist_credit_name WHERE artist = ?) ORDER BY release.date DESC"
        rows, err = server.db.Query(query, artist.id, artist.id)
        if err != nil {
            log.Fatal(err)
        }
        for rows.Next() {
            var release Release
            err = rows.Scan(&release.id, &release.mbid, &release.name, &release.artistCreditId, &release.date, &release._type, &release.coverUrl)
            if err != nil {
                log.Fatal(err)
            }
            release.artistCredit = server.artistCreditFromId(release.artistCreditId)
            artist.appearances = append(artist.appearances, release)
        }
        html := server.generateArtistHTML(artist)
        fmt.Fprint(w, html)
    }
}
func (server Server) generateTrackJSON(track Track) string {
    // manually converts Track object to JSON
    json := "{"
    name := strings.Replace(track.name, "\"", "\\\"", -1)
    json += fmt.Sprintf("\"name\": \"%s\",", name)
    json += fmt.Sprintf("\"number\": %d,", track.number)
    json += fmt.Sprintf("\"length\": %d,", track.length)
    json += fmt.Sprintf("\"url\": \"%s\",", url.PathEscape(track.url))
    json += fmt.Sprintf("\"mbid\": \"%s\",", track.mbid)
    json += fmt.Sprintf("\"releaseCoverUrl\": \"%s\",", url.PathEscape(track.releaseCoverUrl))
    artistCreditHTML := server.generateArtistCreditHTML(track.artistCredit)
    //replace " with \"
    artistCreditHTML = strings.Replace(artistCreditHTML, "\"", "\\\"", -1)
    json += fmt.Sprintf("\"artistCredit\": \"%s\"", artistCreditHTML)
    json += "}"
    return json
}
func trackHandler(w http.ResponseWriter, r *http.Request, server Server) {
    mbid := strings.TrimPrefix(r.URL.Path, "/track/")
    if mbid == "all" {
        html := server.generateAllTrackHTML()
        fmt.Fprint(w, html)
    } else {
        track := Track{}
        query := "SELECT * FROM track WHERE mbid = ?"
        rows, err := server.db.Query(query, mbid)
        for rows.Next() {
            err = rows.Scan(&track.id, &track.mbid, &track.name, &track.number, &track.artistCreditId, &track.length, &track.release, &track.url)
            if err != nil {
                log.Fatal(err)
            }
        }
        track.artistCredit = server.artistCreditFromId(track.artistCreditId)
        // Get release cover url
        query = "SELECT cover_url FROM release WHERE id = ?"
        row := server.db.QueryRow(query, track.release)
        if err := row.Scan(&track.releaseCoverUrl); err != nil {
            log.Printf("No release with id %d", track.release)
        }
        html := server.generateTrackJSON(track)
        fmt.Fprint(w, html)
    }
}
func releaseHandler(w http.ResponseWriter, r *http.Request, server Server) {
    mbid := strings.TrimPrefix(r.URL.Path, "/release/")
    if mbid == "all" {
        html := server.generateAllReleaseHTML()
        fmt.Fprint(w, html)
    } else {
        // Get release from database using mbid
        query := "SELECT * FROM release WHERE mbid = ?"
        row := server.db.QueryRow(query, r.URL.Path[9:])
        release := Release{}
        if err := row.Scan(&release.id, &release.mbid, &release.name, &release.artistCreditId, &release.date, &release._type, &release.coverUrl); err != nil {
            log.Printf("No release with mbid %s", r.URL.Path[9:])
        }
        // Get artist credits from database using artistCreditId
        release.artistCredit = server.artistCreditFromId(release.artistCreditId)
        // Get tracks from database using releaseId
        rows, err := server.db.Query("SELECT * FROM track WHERE release = ? ORDER BY track.number", release.id)
        if err != nil {
            log.Printf("No tracks with release %d", release.id)
        }
        for rows.Next() {
            track := Track{}
            if err := rows.Scan(&track.id, &track.mbid, &track.name, &track.number, &track.artistCreditId, &track.length, &track.release, &track.url); err != nil {
                log.Printf("Error scanning track for release %s", release.mbid)
            }
            release.tracks = append(release.tracks, track)
        }
        // Get artist credits for tracks
        for i, track := range release.tracks {
            release.tracks[i].artistCredit = server.artistCreditFromId(track.artistCreditId)
        }
        // Generate HTML
        html := server.generateReleaseHTML(release)
        fmt.Fprint(w, html)
    }
}
func searchHandler(w http.ResponseWriter, r *http.Request, server Server) {
    searchQuery := r.URL.Query().Get("q")
    // Get releases from database using query
    rows, err := server.db.Query("SELECT * FROM release WHERE name LIKE ?", "%"+searchQuery+"%")
    if err != nil {
        log.Printf("No releases with name %s", searchQuery)
    }
    releases := []Release{}
    for rows.Next() {
        release := Release{}
        if err := rows.Scan(&release.id, &release.mbid, &release.name, &release.artistCreditId, &release.date, &release._type, &release.coverUrl); err != nil {
            log.Printf("Error scanning release for query %s", searchQuery)
        }
        releases = append(releases, release)
    }
    // Get artist credits for each release using artistCreditId
    for i, release := range releases {
        releases[i].artistCredit = server.artistCreditFromId(release.artistCreditId)
    }
    // Get tracks from database using query
    rows, err = server.db.Query("SELECT * FROM track WHERE name LIKE ?", "%"+searchQuery+"%")
    if err != nil {
        log.Printf("No tracks with name %s", searchQuery)
    }
    tracks := []Track{}
    for rows.Next() {
        track := Track{}
        if err := rows.Scan(&track.id, &track.mbid, &track.name, &track.number, &track.artistCreditId, &track.length, &track.release, &track.url); err != nil {
            log.Printf("Error scanning track for query %s", searchQuery)
        }
        tracks = append(tracks, track)
    }
    // Get artist credits for tracks
    for i, track := range tracks {
        tracks[i].artistCredit = server.artistCreditFromId(track.artistCreditId)
    }
    // Get artists_credit_names from database using query
    rows, err = server.db.Query("SELECT * FROM artist_credit_name WHERE name LIKE ? GROUP BY name", "%"+searchQuery+"%" )
    if err != nil {
        log.Printf("No artist_credit_name with name %s", searchQuery)
    }
    artistCreditNames := []ArtistCreditName{}
    for rows.Next() {
        artistCreditName := ArtistCreditName{}
        if err := rows.Scan(&artistCreditName.artistCreditId, &artistCreditName.position, &artistCreditName.artistId, &artistCreditName.name); err != nil {
            log.Printf("Error scanning artist_credit_name for query %s", searchQuery)
        }
        // Get artist credit name mbid
        artistCreditNames = append(artistCreditNames, artistCreditName)
    }
    // Get artistMbid for artist credit names
    for i, artistCreditName := range artistCreditNames {
        row := server.db.QueryRow("SELECT mbid FROM artist WHERE id = ?", artistCreditName.artistId)
        if err := row.Scan(&artistCreditNames[i].artistMbid); err != nil {
            log.Printf("No artist with id %d", artistCreditName.artistId)
        }
    }

    // Generate HTML
    html := server.generateSearchHTML(releases, tracks, artistCreditNames)
    fmt.Fprint(w, html)
}
func main() {
    server := Server{}
    var err error
    server.db, err = sql.Open("sqlite3", "mb.db")
    if err != nil {
        log.Fatal(err)
    }
    server.navHTML, err = os.ReadFile("./static/nav.html")
    server.playerHTML, err = os.ReadFile("./static/player.html")
    server.releaseHTML, err = os.ReadFile("./static/release.html")
    server.smallReleaseHTML, err = os.ReadFile("./static/smallRelease.html")
    server.trackHTML, err = os.ReadFile("./static/track.html")
    server.smallTrackHTML, err = os.ReadFile("./static/smallTrack.html")
    server.artistHTML, err = os.ReadFile("./static/artist.html")
    server.smallArtistHTML, err = os.ReadFile("./static/smallArtist.html")
    server.allTrackHTML, err = os.ReadFile("./static/allTrack.html")
    server.allReleaseHTML, err = os.ReadFile("./static/allRelease.html")
    server.allArtistHTML, err = os.ReadFile("./static/allArtist.html")
    server.searchHTML, err = os.ReadFile("./static/search.html")
    server.headHTML, err = os.ReadFile("./static/head.html")
    http.HandleFunc("/release/", func(w http.ResponseWriter, r *http.Request) {
        releaseHandler(w, r, server)
    })
    http.HandleFunc("/artist/", func(w http.ResponseWriter, r *http.Request) {
        artistHandler(w, r, server)
    })
    http.HandleFunc("/track/", func(w http.ResponseWriter, r *http.Request) {
        trackHandler(w, r, server)
    })
    http.HandleFunc("/write/", func(w http.ResponseWriter, r *http.Request) {
        mbid := r.URL.Path[len("/write/"):]
        server.getNamesFromMBID(mbid)
    })
    http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
        searchHandler(w, r, server)
    })
    mediaFileServer := http.FileServer(http.Dir("/"))
    mediaHandler := http.StripPrefix("/media/", mediaFileServer)
    http.Handle("/media/", mediaHandler)
    http.Handle("/static/", http.FileServer(http.Dir("./")))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
