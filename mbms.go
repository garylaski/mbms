package main

import (
    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "net/http"
    "os"
    "strings"
    "net/http/pprof"
)
type Server struct {
    db               *sql.DB
    navHTML          string
    playerHTML       string
    releaseHTML      string
    artistHTML       string
    releaseItemHTML  string
    trackItemHTML    string
    listHTML         string
    artistItemHTML   string
    headHTML         string
    playlistItemHTML string
    editListHTML     string
    playlistSelectorItemHTML string
    artist_artist_rel_map map[string][2]string
}
func convertMsToTime(ms int) string {
    sec := ms / 1000
    min := sec / 60
    sec = sec % 60
    return fmt.Sprintf("%d:%02d", min, sec)
}
func (server Server) queryRelease(mbid string) (string, string, string, string, string, string)  {
    var id, artist_credit_id, _type_id, name, date, cover_url string
    err := server.db.QueryRow("SELECT rowid,* FROM release WHERE mbid = ?", mbid).Scan(&id, &mbid, &name, &artist_credit_id, &date, &_type_id, &cover_url)
    if err != nil {
        log.Println("queryRelease: ", err)
    }
    return id, name, artist_credit_id, date, _type_id, cover_url
}
func (server Server) generateTrackListHTML(title string, rows *sql.Rows) string {
    var id, mbid, name, number, artist_credit_id, release_id, url, items string
    var length int
    empty := true
    for rows.Next() {
        empty = false
        err := rows.Scan(&id, &mbid, &name, &number, &artist_credit_id, &length, &release_id, &url)
        if err != nil {
            log.Println("generateTrackListHTML: ", err)
        }
        items += server.trackItemHTML
        items = strings.Replace(items, "{{track.name}}", name, 1)
        items = strings.Replace(items, "{{track.id}}", id, 3)
        items = strings.Replace(items, "{{track.number}}", number, 1)
        items = strings.Replace(items, "{{track.length}}", convertMsToTime(length), 1)
        items = strings.Replace(items, "{{track.artistCredit}}", server.generateArtistCreditHTML(artist_credit_id), 1)
    }
    if empty {
        return ""
    }
    html := server.listHTML
    html = strings.Replace(html, "{{title}}", title, 1)
    html = strings.Replace(html, "{{type}}", "track", 1)
    html = strings.Replace(html, "{{items}}", items, 1)
    return html
}
func (server Server) generateDraggableTrackListHTML(title string, rows *sql.Rows) string {
    var id, mbid, name, number, artist_credit_id, release_id, url, items string
    var length int
    empty := true
    for rows.Next() {
        empty = false
        err := rows.Scan(&id, &mbid, &name, &number, &artist_credit_id, &length, &release_id, &url)
        if err != nil {
            log.Println("generateTrackListHTML: ", err)
        }
        items += server.trackItemHTML
        items = strings.Replace(items, "{{track.name}}", name, 1)
        items = strings.Replace(items, "{{track.id}}", id, 2)
        items = strings.Replace(items, "{{track.number}}", number, 1)
        items = strings.Replace(items, "{{track.length}}", convertMsToTime(length), 1)
        items = strings.Replace(items, "{{track.artistCredit}}", server.generateArtistCreditHTML(artist_credit_id), 1)
    }
    if empty {
        return ""
    }
    html := server.listHTML
    html = strings.Replace(html, "{{title}}", title, 1)
    html = strings.Replace(html, "{{type}}", "track", 1)
    html = strings.Replace(html, "{{items}}", items, 1)
    return html
}
func (server Server) generateSearchHTML(search string) string {
    html := "<div id=\"title\" style=\"display:none\">Search results for \"" + search + "\"</div>"
    rows, err := server.db.Query("SELECT rowid,* FROM release WHERE name LIKE ? LIMIT 16", "%"+search+"%")
    if err != nil {
        log.Println("generateSearchHTML: ", err)
    }
    html += server.generateReleaseListHTML("Releases", rows)
    rows, err = server.db.Query("SELECT rowid,* FROM track WHERE name LIKE ? LIMIT 10", "%"+search+"%")
    if err != nil {
        log.Println("generateSearchHTML: ", err)
    }
    html += server.generateTrackListHTML("Tracks", rows)
    rows, err = server.db.Query("SELECT rowid,* FROM artist WHERE rowid IN (SELECT artist FROM artist_credit_name WHERE name LIKE ? GROUP BY artist) LIMIT 10", "%"+search+"%")
    if err != nil {
        log.Println("generateSearchHTML: ", err)
    }
    html += server.generateArtistListHTML("Artists", rows)
    return html
}
func (server Server) queryArtistCredit(id string) string {
    var name string
    err := server.db.QueryRow("SELECT rowid,* FROM artist_credit WHERE rowid = ?", id).Scan(&id, &name)
    if err != nil {
        log.Println("queryArtistCredit: ", err)
    }
    return name
}
func (server Server) generateReleaseHTML(mbid string) string {
    html := server.releaseHTML
    id, name, artist_credit_id, date, _, url := server.queryRelease(mbid)
    html = strings.Replace(html, "{{release.name}}", name, 1)
    html = strings.Replace(html, "{{release.date}}", date, 1)
    html = strings.Replace(html, "{{release.coverUrl}}", url, 1)
    html = strings.Replace(html, "{{release.artistCredit}}", server.generateArtistCreditHTML(artist_credit_id), 1)
    rows, err := server.db.Query("SELECT rowid,* FROM track WHERE release = ? ORDER BY number", id)
    if err != nil {
        log.Println("generateReleaseHTML: ", err)
    }
    var trackHTML, items, number, release_id string
    var length int
    track_ids := "["
    for rows.Next() {
        err = rows.Scan(&id, &mbid, &name, &number, &artist_credit_id, &length, &release_id, &url)
        if err != nil {
            log.Println("generateReleaseHTML: ", err)
        }
        trackHTML = server.trackItemHTML
        trackHTML = strings.Replace(trackHTML, "{{track.name}}", name, 1)
        trackHTML = strings.Replace(trackHTML, "{{track.id}}", id, 3)
        trackHTML = strings.Replace(trackHTML, "{{track.number}}", number, 1)
        trackHTML = strings.Replace(trackHTML, "{{track.length}}", convertMsToTime(length), 1)
        trackHTML = strings.Replace(trackHTML, "{{track.artistCredit}}", server.generateArtistCreditHTML(artist_credit_id), 1)
        items += trackHTML
        track_ids += id + ","
    }
    track_ids = track_ids[:len(track_ids)-1] + "]"
    html = strings.Replace(html, "{{tracks.id}}", track_ids, 2)
    html = strings.Replace(html, "{{release.tracks}}", items, 1)
    return html
}
func (server Server) getTrack(id string) string {
    // json of song
    var number, release_id, length int
    var mbid, name, url, artist_credit_id string
    err := server.db.QueryRow("SELECT rowid,* FROM track WHERE rowid = ?", id).Scan(&id, &mbid, &name, &number, &artist_credit_id, &length, &release_id, &url)
    if err != nil {
        log.Println("getTrack: ", err)
    }
    var cover_url, artist_credit_html, release_mbid string
    err = server.db.QueryRow("SELECT cover_url, mbid FROM release WHERE rowid = ?", release_id).Scan(&cover_url, &release_mbid)
    if err != nil {
        log.Println("getTrack: ", err)
    }
    artist_credit_html = server.generateArtistCreditHTML(artist_credit_id)
    return fmt.Sprintf("{\"id\": %q, \"name\": %q, \"number\": %d, \"artist_credit_id\": %q, \"length\": %d, \"release_id\": %d, \"url\": %q, \"artist_credit_html\": %q, \"cover_url\": %q, \"release_mbid\": %q}", id, name, number, artist_credit_id, length, release_id, url, artist_credit_html, cover_url, release_mbid)
}
func (server Server) generateArtistCreditHTML(id string) string {
    var artist, mbid string
    name := server.queryArtistCredit(id)
    html := name
    rows, err := server.db.Query("SELECT artist, name FROM artist_credit_name WHERE artist_credit = ?", id)
    if err != nil {
        log.Println("generateArtistCreditHTML: ", err)
    }
    for rows.Next() {
        err = rows.Scan(&artist, &name)
        if err != nil {
            log.Println("generateArtistCreditHTML: ", err)
        }
        err = server.db.QueryRow("SELECT mbid FROM artist WHERE rowid = ?", artist).Scan(&mbid)
        if err != nil {
            log.Fatal(err)
        }
        html = strings.Replace(html, name, fmt.Sprintf("<a href=\"/artist/%s\" onclick=\"ajax('/artist/%s'); return false;\">%s</a>", mbid, mbid, name), 1)
    }
    return html
}
func (server Server) generateReleaseListHTML(title string, rows *sql.Rows) string {
    var id, name, mbid, artist_credit_id, date, type_id, cover_url, items string
    empty := true
    for rows.Next() {
        empty = false
        err := rows.Scan(&id, &mbid, &name, &artist_credit_id, &date, &type_id, &cover_url)
        if err != nil {
            log.Println("generateArtistHTML: ", err)
        }
        items += server.releaseItemHTML
        items = strings.Replace(items, "{{release.name}}", name, 2)
        items = strings.Replace(items, "{{release.mbid}}", mbid, 4)
        items = strings.Replace(items, "{{release.date}}", date, 1)
        items = strings.Replace(items, "{{release.coverUrl}}", cover_url, 1)
        items = strings.Replace(items, "{{release.artistCredit}}", server.generateArtistCreditHTML(artist_credit_id), 1)
    }
    if empty {
        return ""
    }
    html := server.listHTML
    html = strings.Replace(html, "{{title}}", title, 1)
    html = strings.Replace(html, "{{type}}", "release", 1)
    html = strings.Replace(html, "{{items}}", items, 1)
    return html
}
func (server Server) getId(table string, selector string, query string) string {
    var id string
    err := server.db.QueryRow(fmt.Sprintf("SELECT rowid FROM %s WHERE %s = ?", table, selector), query).Scan(&id)
    if err != nil {
        log.Println("getId: ", err)
    }
    return id
}
func (server Server) generateArtistListHTML(title string, rows *sql.Rows) string {
    var id, mbid, items string
    var name sql.NullString
    empty := true
    for rows.Next() {
        empty = false
        err := rows.Scan(&id, &mbid, &name)
        if err != nil {
            log.Println("generateArtistListHTML: ", err)
        }
        items += server.artistItemHTML
        items = strings.Replace(items, "{{name}}", name.String, 1)
        items = strings.Replace(items, "{{mbid}}", mbid, 2)
    }
    if empty {
        return ""
    }
    html := server.listHTML
    html = strings.Replace(html, "{{title}}", title, 1)
    html = strings.Replace(html, "{{type}}", "artist", 1)
    html = strings.Replace(html, "{{items}}", items, 1)
    return html
}
func (server Server) generateArtistHTML(mbid string) string {
    html := server.artistHTML
    var artist_id string
    var artist_name sql.NullString
    err := server.db.QueryRow("SELECT rowid,* FROM artist WHERE mbid = ?", mbid).Scan(&artist_id, &mbid, &artist_name)
    if err != nil {
        log.Println("generateArtistHTML: ", err)
    }
    html = strings.Replace(html, "{{artist.name}}", artist_name.String, 1)
    rows, err := server.db.Query("SELECT rowid,* FROM release WHERE artist_credit IN (SELECT artist_credit FROM artist_credit_name WHERE artist = ?) ORDER BY date DESC", artist_id)
    if err != nil {
        log.Println("generateArtistHTML: ", err)
    }
    items := server.generateReleaseListHTML("Releases", rows)
    html = strings.Replace(html, "{{artist.releases}}", items, 1)
    items = ""
    rows, err = server.db.Query("SELECT rowid,* FROM release WHERE rowid IN (SELECT release FROM track WHERE artist_credit IN (SELECT artist_credit FROM artist_credit_name WHERE artist = ?) GROUP BY release) AND release.artist_credit NOT IN (SELECT artist_credit FROM artist_credit_name WHERE artist = ?) ORDER BY release.date DESC", artist_id, artist_id)
    if err != nil {
        log.Println("generateArtistHTML: ", err)
    }
    items = server.generateReleaseListHTML("Appears on", rows)
    html = strings.Replace(html, "{{artist.appearances}}", items, 1)
    rows, err = server.db.Query("SELECT artist,type,direction FROM l_artist_artist WHERE artist = ? GROUP by type", artist_id)
    items = server.generateRelatedArtistListHTML(rows)
    html = strings.Replace(html, "{{artist.artist-rels}}", items, 1)
    return html
}
func (server Server) generateRelatedArtistListHTML(rows *sql.Rows) string {
    var type_id, direction_id int
    var direction bool
    var artist_id, items, type_name string
    empty := true
    for rows.Next() {
        empty = false
        err := rows.Scan(&artist_id, &type_id, &direction)
        if (direction) {
            direction_id = 1
        } else {
            direction_id = 0    
        }
        if err != nil {
            log.Println("generateRelatedArtistListHTML: ", err)
        }
        err = server.db.QueryRow("SELECT name FROM t_artist_artist WHERE rowid = ?", type_id).Scan(&type_name)
        if err != nil {
            log.Println("generateRelatedArtistListHTML: ", err)
        }
        artist_rows, err := server.db.Query("SELECT rowid,* FROM artist WHERE rowid IN (SELECT l_artist FROM l_artist_artist WHERE artist = ? AND type = ?)", artist_id, type_id)
        if err != nil {
            log.Println("generateRelatedArtistListHTML: ", err)
        }
        type_name = server.artist_artist_rel_map[type_name][direction_id]
        items += server.generateArtistListHTML(type_name, artist_rows)
    }
    if empty {
        return ""
    }
    return items
}
func (server Server) updatePlaylist(id string, tracks string, remove string) string {
    query := "DELETE FROM playlist_track WHERE playlist = ? AND rowid IN (" + remove + ")"
    _, err := server.db.Exec(query, id)
    if err != nil {
        log.Println("updatePlaylist: ", err)
    }
    for _, track := range strings.Split(tracks, ",") {
        query = "INSERT INTO playlist_track (playlist, track) VALUES (?, ?)"
        _, err := server.db.Exec(query, id, track)
        if err != nil {
            log.Println("updatePlaylist: ", err)
        }
    }
    return "OK"
}
func (server Server) createPlaylist(name string) string {
    query := "INSERT INTO playlist (name) VALUES (?)"
    _, err := server.db.Exec(query, name)
    if err != nil {
        log.Println("createPlaylist: ", err)
    }
    return "OK"
}
func (server Server) generatePlaylistsHTML() string {
    query := "SELECT rowid,name FROM playlist"
    rows, err := server.db.Query(query)
    if err != nil {
        log.Println("generatePlaylistsHTML: ", err)
    }
    var id, name, items string
    for rows.Next() {
        err := rows.Scan(&id, &name)
        if err != nil {
            log.Println("generatePlaylistsHTML: ", err)
        }
        items += server.playlistItemHTML
        items = strings.Replace(items, "{{name}}", name, 1)
        items = strings.Replace(items, "{{id}}", id, 3)
    }
    html := server.editListHTML
    html = strings.Replace(html, "{{title}}", "Playlists", 1)
    html = strings.Replace(html, "{{type}}", "playlist", 1)
    html = strings.Replace(html, "{{items}}", items, 1)
    return html
}
func (server Server) generatePlaylistHTML(id string) string {
    query := "SELECT name FROM playlist WHERE rowid = ?"
    var name string
    err := server.db.QueryRow(query, id).Scan(&name)
    if err != nil {
        log.Println("generatePlaylistHTML: ", err)
    }
    query = "SELECT rowid,* FROM track WHERE rowid IN (SELECT track FROM playlist_track WHERE playlist = ?)"
    rows, err := server.db.Query(query, id)
    if err != nil {
        log.Println("generatePlaylistHTML: ", err)
    }
    return server.generateTrackListHTML(name, rows)
}
func (server Server) generateReleasesHTML() string {
    query := "SELECT rowid,* FROM release ORDER BY RANDOM() LIMIT 50"
    rows, err := server.db.Query(query)
    if err != nil {
        log.Println("generateReleasesHTML: ", err)
    }
    return server.generateReleaseListHTML("Releases", rows)
}
func (server Server) generatePage(content string) string {
    html := "<html>"
    html += string(server.headHTML)
    html += "<body>"
    html += string(server.navHTML)
    html += "<div id=\"content\">"
    html += content
    html += "</div>"
    html += string(server.playerHTML)
    html += "</body>"
    html += "</html>"
    return html
}
func fsCache(fs http.Handler) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Cache-Control", "public, max-age=31536000")
        fs.ServeHTTP(w, r)
    }
}
func (server Server) deletePlaylist(id string) string {
    query := "DELETE FROM playlist WHERE rowid = ?"
    _, err := server.db.Exec(query, id)
    if err != nil {
        log.Println("deletePlaylist: ", err)
    }
    query = "DELETE FROM playlist_track WHERE playlist = ?"
    _, err = server.db.Exec(query, id)
    if err != nil {
        log.Println("deletePlaylist: ", err)
    }
    return "OK"
}
func (server Server) generatePlaylistSelectorHTML(rows* sql.Rows, tid string) string {
    var id, name, items string
    for rows.Next() {
        err := rows.Scan(&id, &name)
        if err != nil {
            log.Println("generatePlaylistSelectorHTML: ", err)
        }
        items += server.playlistSelectorItemHTML
        items = strings.Replace(items, "{{name}}", name, 1)
        items = strings.Replace(items, "{{pid}}", id, 1)
        items = strings.Replace(items, "{{tid}}", tid, 1)
    }
    html := server.listHTML
    html = strings.Replace(html, "{{title}}", "Playlists", 1)
    html = strings.Replace(html, "{{type}}", "playlist-chooser", 1)
    html = strings.Replace(html, "{{items}}", items, 1)
    return html
}
func (server Server) getPlaylists(format string, track string) string {
    query := "SELECT rowid,name FROM playlist"
    rows, err := server.db.Query(query)
    if err != nil {
        log.Println("getPlaylists: ", err)
    }
    if format == "html" {
        return server.generatePlaylistSelectorHTML(rows, track)
    }
    return "OK"
}
func main() {
    server := Server{}
    mux := http.NewServeMux()
    var err error
    server.db, err = sql.Open("sqlite3", "file:mb.db?cache=shared")
    if err != nil {
        log.Fatal(err)
    }
    navHTML, err := os.ReadFile("./static/nav.html")
    playerHTML, err := os.ReadFile("./static/player.html")
    releaseHTML, err := os.ReadFile("./static/release.html")
    trackItemHTML, err := os.ReadFile("./static/track-item.html")
    headHTML, err := os.ReadFile("./static/head.html")
    releaseItemHTML, err := os.ReadFile("./static/release-item.html")
    artistHTML, err := os.ReadFile("./static/artist.html")
    listHTML, err := os.ReadFile("./static/list.html")
    artistItemHTML, err := os.ReadFile("./static/artist-item.html")
    playlistItemHTML, err := os.ReadFile("./static/playlist-item.html")
    editListHTML, err := os.ReadFile("./static/edit-list.html")
    playlistSelectorItemHTML, err := os.ReadFile("./static/playlist-selector-item.html")
    server.playlistSelectorItemHTML = string(playlistSelectorItemHTML)
    server.editListHTML = string(editListHTML)
    server.playlistItemHTML = string(playlistItemHTML)
    server.navHTML = string(navHTML)
    server.playerHTML = string(playerHTML)
    server.releaseHTML = string(releaseHTML)
    server.trackItemHTML = string(trackItemHTML)
    server.headHTML = string(headHTML)
    server.releaseItemHTML = string(releaseItemHTML)
    server.artistHTML = string(artistHTML)
    server.listHTML = string(listHTML)
    server.artistItemHTML = string(artistItemHTML)
    server.artist_artist_rel_map = map[string][2]string{
        "member of band": [2]string{"Members:", "Member of:"},
        "is person": [2]string{"Legal name:", "Also performs as:"},
        "involved with": [2]string{"Involved with:", "Involved with:"},
        "founder": [2]string{"Founded:", "Founded by:"},
        "sibling": [2]string{"Siblings:", "Siblings:"},
        "parent": [2]string{"Parents:", "Children:"},
    }
    mux.HandleFunc("/release/", func(w http.ResponseWriter, r *http.Request) {
        mbid := r.URL.Path[9:]
        fmt.Fprint(w, server.generatePage(server.generateReleaseHTML(mbid)))
    })
    mux.HandleFunc("/artist/", func(w http.ResponseWriter, r *http.Request) {
        mbid := r.URL.Path[8:]
        fmt.Fprint(w, server.generatePage(server.generateArtistHTML(mbid)))
    })
    mux.HandleFunc("/releases", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, server.generatePage(server.generateReleasesHTML()))
    })
    mux.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
        query := r.URL.Query().Get("q")
        fmt.Fprint(w, server.generatePage(server.generateSearchHTML(query)))
    })
    mux.HandleFunc("/playlist/", func(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Path[10:]
        fmt.Fprint(w, server.generatePage(server.generatePlaylistHTML(id)))
    })
    mux.HandleFunc("/playlists", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, server.generatePage(server.generatePlaylistsHTML()))
    })
    mux.HandleFunc("/ajax/playlist/", func(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Path[15:]
        fmt.Fprint(w, server.generatePlaylistHTML(id))
    })
    mux.HandleFunc("/ajax/playlists", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, server.generatePlaylistsHTML())
    })
    mux.HandleFunc("/ajax/release/", func(w http.ResponseWriter, r *http.Request) {
        mbid := r.URL.Path[14:]
        fmt.Fprint(w, server.generateReleaseHTML(mbid))
    })
    mux.HandleFunc("/ajax/releases", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, server.generateReleasesHTML())
    })
    mux.HandleFunc("/ajax/artist/", func(w http.ResponseWriter, r *http.Request) {
        mbid := r.URL.Path[13:]
        fmt.Fprint(w, server.generateArtistHTML(mbid))
    })
    mux.HandleFunc("/ajax/search", func(w http.ResponseWriter, r *http.Request) {
        query := r.URL.Query().Get("q")
        fmt.Fprint(w, server.generateSearchHTML(query))
    })
    mux.HandleFunc("/rest/getTrack", func(w http.ResponseWriter, r *http.Request) {
        id := r.URL.Query().Get("id")
        fmt.Fprint(w, server.getTrack(id))
    })
    mux.HandleFunc("/rest/getPlaylists", func(w http.ResponseWriter, r *http.Request) {
        err := r.ParseForm()
        if err != nil {
            log.Fatal(err)
        }
        format := r.Form.Get("fmt")
        id := r.Form.Get("track")
        fmt.Fprint(w, server.getPlaylists(format, id))
    })

    mux.HandleFunc("/rest/updatePlaylist", func(w http.ResponseWriter, r *http.Request) {
        err := r.ParseForm()
        if err != nil {
            log.Fatal(err)
        }
        id := r.Form.Get("playlistId")
        tracks := r.Form.Get("songIdToAdd")
        remove := r.Form.Get("songIndexToRemove")
        fmt.Fprint(w, server.updatePlaylist(id, tracks, remove))
    })
    mux.HandleFunc("/rest/createPlaylist", func(w http.ResponseWriter, r *http.Request) {
        err := r.ParseForm()
        if err != nil {
            log.Fatal(err)
        }
        name := r.Form.Get("name")
        fmt.Fprint(w, server.createPlaylist(name))
    })
    mux.HandleFunc("/rest/deletePlaylist", func(w http.ResponseWriter, r *http.Request) {
        err := r.ParseForm()
        if err != nil {
            log.Fatal(err)
        }
        id := r.Form.Get("id")
        log.Println(id)
        fmt.Fprint(w, server.deletePlaylist(id))
    })

    mux.Handle("/media/", http.StripPrefix("/media/", fsCache(http.FileServer(http.Dir("/")))))
    mux.Handle("/static/", fsCache(http.FileServer(http.Dir("./"))))
    mux.Handle("/favicon.ico", fsCache(http.FileServer(http.Dir("./static"))))
    // profiling routes
    mux.HandleFunc("/debug/pprof/", pprof.Index)
    mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
    mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
    mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
    mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

    mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
    mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
    mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
    mux.Handle("/debug/pprof/block", pprof.Handler("block"))

    log.Fatal(http.ListenAndServe(":8080", mux))
}
