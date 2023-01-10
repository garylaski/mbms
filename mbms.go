package main
import (
    "io"
    "encoding/json"
    "os"
    "database/sql"
    "fmt"
    "net/url"
    "log"
    "sort"
    "strconv"
    "strings"
    "net/http"
    _ "github.com/mattn/go-sqlite3"
)
type Artist struct {
    id int
    mbid string
    releases []Release
    name string
}
type ArtistCredit struct {
    id int
    name string
    artistCount int
    artistCreditNames []ArtistCreditName
}
type ArtistCreditName struct {
    artistCreditId int
    position int
    name string
    artistId int
    artistMbid string
}
type Release struct {
    id int
    mbid string
    name string
    artistCreditId int
    date int
    _type int
    coverUrl string
    tracks []Track
    artistCredit ArtistCredit
}
type Track struct {
    id int
    mbid string
    name string
    number int
    artistCreditId int
    artistCredit ArtistCredit
    length int
    release int
    url string
}
type Server struct {
    db *sql.DB
    navHTML []byte
    playerHTML []byte
    releaseHTML []byte
    trackHTML []byte
    artistHTML []byte
    smallReleaseHTML []byte
    smallTrackHTML []byte
    smallArtistHTML []byte
    allReleaseHTML []byte
    allTrackHTML []byte
    allArtistHTML []byte
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
    html := string(server.releaseHTML)
    html = strings.Replace(html, "{{nav}}", string(server.navHTML), -1)
    html = strings.Replace(html, "{{player}}", string(server.playerHTML), -1)
    html = strings.Replace(html, "{{release.name}}", release.name, -1)
    html = strings.Replace(html, "{{release.date}}", strconv.Itoa(release.date), -1)
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
    html = strings.Replace(html, "{{release.date}}", strconv.Itoa(release.date), -1)
    html = strings.Replace(html, "{{release.coverUrl}}", url.PathEscape(release.coverUrl), -1)
    html = strings.Replace(html, "{{release.artistCredit}}", server.generateArtistCreditHTML(release.artistCredit), -1)
    return html
}
func (server Server) generateArtistHTML(artist Artist) string {
    html := string(server.artistHTML)
    html = strings.Replace(html, "{{nav}}", string(server.navHTML), -1)
    html = strings.Replace(html, "{{player}}", string(server.playerHTML), -1)
    if (artist.name == "") {
        artist.name = server.getNamesFromMBID(artist.mbid)
    }
    html = strings.Replace(html, "{{artist.name}}", artist.name, -1)
    var releasesHTML = ""
    for _, release := range artist.releases {
        releasesHTML += server.generateSmallReleaseHTML(release)
    }
    html = strings.Replace(html, "{{artist.releases}}", releasesHTML, -1)
    return html
}
func (server Server) generateTrackHTML(track Track) string {
    html := string(server.trackHTML)
    return html
}
func (server Server) generateArtistCreditHTML(artistCredit ArtistCredit) string {
    artistCreditNameHTML := artistCredit.name
    for _, artistCreditName := range artistCredit.artistCreditNames {
        artistCreditNameHTML = strings.Replace(artistCreditNameHTML, artistCreditName.name, fmt.Sprintf("<a href=\"../../artist/%s\" onclick=\"ajax('../../artist/%s'); return false;\">%s</a>", artistCreditName.artistMbid, artistCreditName.artistMbid, artistCreditName.name), -1)
    }
    return artistCreditNameHTML

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
func artistHandler(w http.ResponseWriter, r *http.Request, server Server) {
    mbid := strings.TrimPrefix(r.URL.Path, "/artist/")
    if (mbid == "all") {
        html := server.generateAllArtistHTML()
        fmt.Fprint(w, html)
    } else {
        artist := Artist{}
        query := "SELECT * FROM artist WHERE mbid = ?"
        rows, err := server.db.Query(query, mbid)
        for rows.Next() {
            err = rows.Scan(&artist.id, &artist.mbid, &artist.name)
        }
        if err != nil {
            fmt.Println(err)
        }
        query = "SELECT * FROM artist_credit WHERE id IN (SELECT artist_credit FROM artist_credit_name WHERE artist = ?)"
        rows, err = server.db.Query(query, artist.id)
        if err != nil {
            log.Fatal(err)
        }
        var artist_credits []ArtistCredit
        for rows.Next() {
            var artist_credit ArtistCredit
            err = rows.Scan(&artist_credit.id, &artist_credit.name, &artist_credit.artistCount)
            if err != nil {
                log.Fatal(err)
            }
            artist_credits = append(artist_credits, artist_credit)
        }
        //get releases from artist credits id
        for _, artist_credit := range artist_credits {
            query = "SELECT * FROM release WHERE artist_credit = ?"
            rows, err = server.db.Query(query, artist_credit.id)
            if err != nil {
                log.Fatal(err)
            }
            for rows.Next() {
                var release Release
                err = rows.Scan(&release.id, &release.mbid, &release.name, &release.artistCreditId, &release.date, &release._type, &release.coverUrl)
                release.artistCredit = artist_credit
                if err != nil {
                    log.Fatal(err)
                }
                // get arist credit names
                rows, err := server.db.Query("SELECT * FROM artist_credit_name WHERE artist_credit = ?", release.artistCreditId)
                if err != nil {
                    log.Printf("No artist_credit_name with artist_credit %d", release.artistCreditId)
                }
                for rows.Next() {
                    artistCreditName := ArtistCreditName{}
                    if err := rows.Scan(&artistCreditName.artistCreditId, &artistCreditName.position, &artistCreditName.artistId, &artistCreditName.name); err != nil {
                        log.Printf("Error scanning artist_credit_name")
                    }
                    release.artistCredit.artistCreditNames = append(release.artistCredit.artistCreditNames, artistCreditName)
                }
                // Get mbid for artist credit names
                query = "SELECT mbid FROM artist WHERE id = ?"
                for i, artistCreditName := range release.artistCredit.artistCreditNames {
                    row := server.db.QueryRow(query, artistCreditName.artistId)
                    if err := row.Scan(&release.artistCredit.artistCreditNames[i].artistMbid); err != nil {
                        log.Printf("No artist with id %d", artistCreditName.artistId)
                    }
                }

                artist.releases = append(artist.releases, release)
            }
        }
        sort.Slice(artist.releases, func(i, j int) bool {
            return artist.releases[i].date > artist.releases[j].date
        })
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
    artistCreditHTML := server.generateArtistCreditHTML(track.artistCredit)
    //replace " with \"
    artistCreditHTML = strings.Replace(artistCreditHTML, "\"", "\\\"", -1)
    json += fmt.Sprintf("\"artistCredit\": \"%s\"", artistCreditHTML)
    json += "}"
    return json
}
func trackHandler(w http.ResponseWriter, r *http.Request, server Server) {
    mbid := strings.TrimPrefix(r.URL.Path, "/track/")
    if (mbid == "all") {
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
        query = "SELECT * FROM artist_credit WHERE id = ?"
        row := server.db.QueryRow(query, track.artistCreditId)
        if err := row.Scan(&track.artistCredit.id, &track.artistCredit.name, &track.artistCredit.artistCount); err != nil {
            log.Printf("No artist_credit with id %d", track.artistCreditId)
        }
        // Get artist credit names from database using artistCreditId
        rows, err = server.db.Query("SELECT * FROM artist_credit_name WHERE artist_credit = ?", track.artistCreditId)
        if err != nil {
            log.Printf("No artist_credit_name with artist_credit %d", track.artistCreditId)
        }
        for rows.Next() {
            artistCreditName := ArtistCreditName{}
            if err := rows.Scan(&artistCreditName.artistCreditId, &artistCreditName.position, &artistCreditName.artistId, &artistCreditName.name); err != nil {
                log.Printf("Error scanning artist_credit_name")
            }
            track.artistCredit.artistCreditNames = append(track.artistCredit.artistCreditNames, artistCreditName)
        }
        // Get mbid for artist credit names
        query = "SELECT mbid FROM artist WHERE id = ?"
        for i, artistCreditName := range track.artistCredit.artistCreditNames {
            row := server.db.QueryRow(query, artistCreditName.artistId)
            if err := row.Scan(&track.artistCredit.artistCreditNames[i].artistMbid); err != nil {
                log.Printf("No artist with id %d", artistCreditName.artistId)
            }
        }
        html := server.generateTrackJSON(track)
        fmt.Fprint(w, html)
    }
}
func releaseHandler(w http.ResponseWriter, r *http.Request, server Server) {
    mbid := strings.TrimPrefix(r.URL.Path, "/release/")
    if (mbid == "all") {
        html := server.generateAllReleaseHTML()
        fmt.Fprint(w, html)
    } else  {
        // Get release from database using mbid
        query := "SELECT * FROM release WHERE mbid = ?"
        row := server.db.QueryRow(query, r.URL.Path[9:])
        release := Release{}
        if err := row.Scan(&release.id, &release.mbid, &release.name, &release.artistCreditId, &release.date, &release._type, &release.coverUrl); err != nil {
            log.Printf("No release with mbid %s", r.URL.Path[9:])
        }
        // Get artist credits from database using artistCreditId
        query = "SELECT * FROM artist_credit WHERE id = ?"
        row = server.db.QueryRow(query, release.artistCreditId)
        if err := row.Scan(&release.artistCredit.id, &release.artistCredit.name, &release.artistCredit.artistCount); err != nil {
            log.Printf("No artist credit with id %d", release.artistCreditId)
        }
        // Get artist credit names from database using artistCreditId
        rows, err := server.db.Query("SELECT * FROM artist_credit_name WHERE artist_credit = ?", release.artistCreditId)
        if err != nil {
            log.Printf("No artist_credit_name with artist_credit %d", release.artistCreditId)
        }
        for rows.Next() {
            artistCreditName := ArtistCreditName{}
            if err := rows.Scan(&artistCreditName.artistCreditId, &artistCreditName.position, &artistCreditName.artistId, &artistCreditName.name); err != nil {
                log.Printf("Error scanning artist_credit_name")
            }
            release.artistCredit.artistCreditNames = append(release.artistCredit.artistCreditNames, artistCreditName)
        }
        // Get mbid for artist credit names
        query = "SELECT mbid FROM artist WHERE id = ?"
        for i, artistCreditName := range release.artistCredit.artistCreditNames {
            row := server.db.QueryRow(query, artistCreditName.artistId)
            if err := row.Scan(&release.artistCredit.artistCreditNames[i].artistMbid); err != nil {
                log.Printf("No artist with id %d", artistCreditName.artistId)
            }
        }
        // Get tracks from database using releaseId
        rows, err = server.db.Query("SELECT * FROM track WHERE release = ?", release.id)
        if err != nil {
            log.Printf("No tracks with release %d", release.id)
        }
        for rows.Next() {
            track := Track{}
            if err := rows.Scan(&track.id, &track.mbid, &track.name, &track.number, &track.artistCreditId, &track.length, &track.release, &track.url); err != nil {
                log.Printf("Error scanning track for release %d", release.mbid)
            }
            release.tracks = append(release.tracks, track)
        }
        // Get artist credits for tracks
        for i, track := range release.tracks {
            query = "SELECT * FROM artist_credit WHERE id = ?"
            row = server.db.QueryRow(query, track.artistCreditId)
            if err := row.Scan(&release.tracks[i].artistCredit.id, &release.tracks[i].artistCredit.name, &release.tracks[i].artistCredit.artistCount); err != nil {
                log.Printf("No artist credit with id %d", track.artistCreditId)
            }
            // Get artist credit names from database using artistCreditId
            rows, err := server.db.Query("SELECT * FROM artist_credit_name WHERE artist_credit = ?", track.artistCreditId)
            if err != nil {
                log.Printf("No artist_credit_name with artist_credit %d", track.artistCreditId)
            }
            for rows.Next() {
                artistCreditName := ArtistCreditName{}
                if err := rows.Scan(&artistCreditName.artistCreditId, &artistCreditName.position, &artistCreditName.artistId, &artistCreditName.name); err != nil {
                    log.Printf("Error scanning artist_credit_name")
                }
                release.tracks[i].artistCredit.artistCreditNames = append(release.tracks[i].artistCredit.artistCreditNames, artistCreditName)
            }
            // Get mbid for artist credit names
            query := "SELECT mbid FROM artist WHERE id = ?"
            for j, artistCreditName := range release.tracks[i].artistCredit.artistCreditNames {
                row := server.db.QueryRow(query, artistCreditName.artistId)
                if err := row.Scan(&release.tracks[i].artistCredit.artistCreditNames[j].artistMbid); err != nil {
                    log.Printf("No artist with id %d", artistCreditName.artistId)
                }
            }
        }
        // Sort tracks by number
        sort.Slice(release.tracks, func(i, j int) bool {
            return release.tracks[i].number < release.tracks[j].number
        })
        // Generate HTML
        html := server.generateReleaseHTML(release)
        fmt.Fprint(w, html)
    }
}
func main() {
    server := Server{}
    var err error
    server.db, err = sql.Open("sqlite3", "mb.db")
    if err != nil {
        log.Fatal(err)
    }
    server.navHTML, err  = os.ReadFile("./static/nav.html")
    if err != nil {
        log.Fatal(err)
    }
    server.playerHTML, err  = os.ReadFile("./static/player.html")
    if err != nil {
        log.Fatal(err)
    }
    server.releaseHTML, err = os.ReadFile("./static/release.html")
    if err != nil {
        log.Fatal(err)
    }
    server.smallReleaseHTML, err = os.ReadFile("./static/smallRelease.html")
    if err != nil {
        log.Fatal(err)
    }
    server.trackHTML, err = os.ReadFile("./static/track.html")
    if err != nil {
        log.Fatal(err)
    }
    server.smallTrackHTML, err = os.ReadFile("./static/smallTrack.html")
    if err != nil {
        log.Fatal(err)
    }
    server.artistHTML, err = os.ReadFile("./static/artist.html")
    if err != nil {
        log.Fatal(err)
    }
    server.smallArtistHTML, err = os.ReadFile("./static/smallArtist.html")
    if err != nil {
        log.Fatal(err)
    }
    server.allTrackHTML, err = os.ReadFile("./static/allTrack.html")
    if err != nil {
        log.Fatal(err)
    }
    server.allReleaseHTML, err = os.ReadFile("./static/allRelease.html")
    if err != nil {
        log.Fatal(err)
    }
    server.allArtistHTML, err = os.ReadFile("./static/allArtist.html")
    if err != nil {
        log.Fatal(err)
    }
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
    mediaFileServer := http.FileServer(http.Dir("/"))
    mediaHandler := http.StripPrefix("/media/", mediaFileServer)
    http.Handle("/media/", mediaHandler)
    http.Handle("/static/", http.FileServer(http.Dir("./")))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
