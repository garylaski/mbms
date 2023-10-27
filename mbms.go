package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
    "bytes"
)

type Ac struct {
	mbid string
	id   int
}

type Server struct {
	cache                 map[string]struct{}
	urls                  map[string]string
	artistCache           sync.Map
	artistCreditCache     sync.Map
	artistCreditNameCache sync.Map
	templates             map[string]string
	db                    *pgxpool.Pool
	wg                    sync.WaitGroup
	ctx                   context.Context
    batch                *pgx.Batch
    artistCreditId       int
    bufferPool           sync.Pool
    ffprobePool          sync.Pool
}

type handler func(string) (string, error)
type formatter func(string) string

func ConvertMsToTime(ms int) string {
	sec := ms / 1000
	min := sec / 60
	sec = sec % 60
	return fmt.Sprintf("%d:%02d", min, sec)
}

func (server *Server) Handle(w http.ResponseWriter, r *http.Request, ps httprouter.Params, inner handler, outer formatter) {
	mbid := ps.ByName("mbid")
	var (
		content string
		err     error
	)
	if _, ok := server.cache[mbid]; !ok {
		content, err = inner(mbid)
		if err != nil {
			log.Printf("inner (%v): %s\n", inner, err)
			http.Error(w, "Invalid MBID", http.StatusBadRequest)
			return
		}
		err = os.WriteFile("static/"+mbid+".html", []byte(content), 0644)
		if err != nil {
			log.Printf("WriteFile: %s\n", err)
		} else {
			server.cache[mbid] = struct{}{}
		}
	} else {
		b, err := os.ReadFile("static/" + mbid + ".html")
		if err != nil {
			log.Printf("ReadFile: %s\n", err)
			http.Error(w, "File unavailable.", http.StatusBadRequest)
			return
		} else {
			content = string(b)
		}
	}
	fmt.Fprint(w, outer(content))
}

func (server *Server) HandleRelease(mbid string) (string, error) {
	var (
		name, coverUrl, date string
		artistCreditId       int
	)
	err := server.db.QueryRow(server.ctx,
		"SELECT name,cover_url,date,artist_credit FROM mbms.release WHERE mbid = $1",
		mbid).Scan(&name, &coverUrl, &date, &artistCreditId)
	if err != nil {
		return "", err
	}
	rows, err := server.db.Query(server.ctx,
		"SELECT mbid,name,number,url,length,artist_credit FROM mbms.track WHERE release_mbid = $1 ORDER BY number",
		mbid)
	if err != nil {
		return "", err
	}
    trackList, trackMbids := server.GenerateTrackList(rows)
	server.urls[mbid] = coverUrl
	if coverUrl == "" {
		coverUrl = "static/missing.png"
	} else {
		coverUrl = "media/" + mbid
	}
	return fmt.Sprintf(server.templates["release.html"],
		coverUrl,
		name,
		server.GenerateArtistCredit(artistCreditId),
		date,
		trackMbids,
		trackList), nil
}

func (server *Server) GenerateTrackMbids(rows pgx.Rows) string {
    var (
        list, mbid string
    )
    for rows.Next() {
        err := rows.Scan(&mbid, nil, nil, nil, nil, nil)
        if err != nil {
            log.Printf("GenerateTrackMbids(81): %s\n", err)
        }
        list += mbid + ","
    }
    return list
}

func (server *Server) GenerateArtistCredit(id int) string {
	idString := fmt.Sprintf("%d", id)
	if _, ok := server.cache[idString]; !ok {
		var (
			artistCredit, artistMbid string
			name                     string
		)
		err := server.db.QueryRow(server.ctx,
			"SELECT name FROM mbms.artist_credit WHERE id = $1",
			id).Scan(&artistCredit)
		if err != nil {
			log.Printf("GenerateArtistCredit(135): %s\n", err)
		}
		rows, err := server.db.Query(server.ctx,
			"SELECT name,artist_mbid FROM mbms.artist_credit_name WHERE artist_credit = $1",
			id)
		if err != nil {
			log.Printf("GenerateArtistCredit(141): %s\n", err)
		}
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&name, &artistMbid)
			if err != nil {
				log.Printf("GenerateArtistCredit(147): %s\n", err)
			}
			artistCredit = strings.Replace(artistCredit, name,
				fmt.Sprintf(server.templates["artist-link.html"],
					artistMbid,
					name),
				1)
		}
		server.cache[idString] = struct{}{}
		err = os.WriteFile("static/"+idString+".html", []byte(artistCredit), 0644)
		if err != nil {
			log.Printf("GenerateArtistCredit WriteFile: %s\n", err)
		}
		return artistCredit
	}
	artistCredit, err := os.ReadFile("static/" + idString + ".html")
	if err != nil {
		log.Printf("GenerateArtistCredit ReadFile: %s\n", err)
	}
	return string(artistCredit)
}

func (server *Server) GenerateReleaseList(rows pgx.Rows) string {
	var (
		releases                                 string
		releaseName, coverUrl, releaseMbid, date string
		artistCreditId                           int
	)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&releaseMbid, &releaseName, &coverUrl, &date, &artistCreditId)
		if err != nil {
			log.Printf("GenerateReleaseList Scan rows: %s\n", err)
		}
		server.urls[releaseMbid], err = url.QueryUnescape(coverUrl)
		if err != nil {
			log.Printf("GenerateReleaseList QueryUnescape: %s\n", err)
		}
		// TODO: move cover missing logic to media handler
		if coverUrl == "" {
			coverUrl = "static/missing.png"
		} else {
			coverUrl = "media/" + releaseMbid
		}
		releases += fmt.Sprintf(server.templates["release-li.html"],
			releaseMbid,
			coverUrl,
			releaseName,
			server.GenerateArtistCredit(artistCreditId),
			date)
	}
	return releases
}

func (server *Server) GenerateTrackList(rows pgx.Rows) (string, string) {
    var (
        trackList, mbidList, url, mbid, name string
        artistCreditId, number, length int
    )
    defer rows.Close()
    for rows.Next() {
        err := rows.Scan(&mbid, &name, &number, &url, &length, &artistCreditId)
        if err != nil {
            log.Printf("GenerateTrackList Scan rows: %s\n", err)
        }
        trackList += fmt.Sprintf(server.templates["track-li.html"],
            number,
            name,
            server.GenerateArtistCredit(artistCreditId),
            mbid,
            ConvertMsToTime(length))
        server.urls[mbid] = url
        mbidList += mbid + ","
    }
    return trackList, mbidList
}

func (server *Server) GenerateArtistList(rows pgx.Rows) string {
    var (
        artists, artistMbid, artistName string
    )
    defer rows.Close()
    for rows.Next() {
        err := rows.Scan(&artistMbid, &artistName)
        if err != nil {
            log.Printf("GenerateArtistList Scan rows: %s\n", err)
        }
        artists += fmt.Sprintf(server.templates["artist-li.html"], fmt.Sprintf(server.templates["artist-link.html"], artistMbid, artistName))
    }
    return artists
}

func (server *Server) HandleArtist(mbid string) (string, error) {
	// query database for artist info
	var (
		artistName string
	)
	// get artist_name, id
	err := server.db.QueryRow(server.ctx,
		"SELECT name FROM mbms.artist WHERE mbid = $1",
		mbid).Scan(&artistName)
	if err != nil {
		return "", err
	}
	// get release_names
	rows, err := server.db.Query(server.ctx,
		"SELECT mbid,name,cover_url,date,artist_credit FROM mbms.release WHERE artist_credit IN (SELECT artist_credit FROM mbms.artist_credit_name WHERE artist_mbid = $1) ORDER BY date DESC",
		mbid)
	if err != nil {
		return "", err
	}
    releases := server.GenerateReleaseList(rows)
	rows, err = server.db.Query(server.ctx,
		"SELECT mbid,name,cover_url,date,artist_credit FROM mbms.release WHERE mbid IN (SELECT release_mbid FROM mbms.track WHERE artist_credit IN (SELECT artist_credit FROM mbms.artist_credit_name WHERE artist_mbid = $1) GROUP BY release_mbid) AND release.artist_credit NOT IN (SELECT artist_credit FROM mbms.artist_credit_name WHERE artist_mbid = $1) ORDER BY release.date DESC",
		mbid)
	if err != nil {
		return "", err
	}
	appearences := server.GenerateReleaseList(rows)
	return fmt.Sprintf(server.templates["artist.html"],
		artistName,
		releases,
		appearences), nil
}
func (server *Server) HandleTrack(mbid string) (string, error) {
	var (
		trackName, trackUrl                      string
		trackNumber, trackLength, artistCreditId int
		releaseMbid                              string
	)
	err := server.db.QueryRow(server.ctx,
		"SELECT name,number,url,length,release_mbid,artist_credit FROM mbms.track WHERE mbid = $1",
		mbid).Scan(&trackName, &trackNumber, &trackUrl, &trackLength, &releaseMbid, &artistCreditId)
	if err != nil {
		return "", err
	}
	server.urls[mbid] = trackUrl

	return fmt.Sprintf(server.templates["track.json"],
		trackName,
		mbid,
		trackNumber,
		trackLength,
		server.GenerateArtistCredit(artistCreditId),
		releaseMbid), nil
}
func (server *Server) LoadTemplate(name string) {
	b, err := os.ReadFile("templates/" + name)
	if err != nil {
		log.Fatal(err)
	}
	server.templates[name] = string(b)
}
func (server *Server) HandleMedia(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	mbid := ps.ByName("mbid")
	if _, ok := server.urls[mbid]; !ok {
		log.Printf("Function: %s\n", "Invalid MBID")
		http.Error(w, "Invalid MBID", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, server.urls[mbid])
}
func (server *Server) HandleSearch(query string) string {
	if query == "" {
		log.Printf("HandleSearch: %s\n", "Empty query")
		return fmt.Sprintf(server.templates["search.html"], "", "")
	}
	releaseRows, err := server.db.Query(server.ctx,
		"SELECT mbid,name,cover_url,date,artist_credit FROM mbms.release WHERE name ILIKE $1",
		"%"+query+"%")
	if err != nil {
		log.Printf("HandleSearch: %s\n", "Release query error")
		return fmt.Sprintf(server.templates["search.html"], "", "")
	}
    trackRows, err := server.db.Query(server.ctx,
        "SELECT mbid,name,number,url,length,artist_credit FROM mbms.track WHERE name ILIKE $1",
        "%"+query+"%")
    if err != nil {
        log.Printf("HandleSearch: %s, %v\n", "Track query error", err)
        return fmt.Sprintf(server.templates["search.html"], "", "")
    }
    artistRows, err := server.db.Query(server.ctx,
        "SELECT mbid,name FROM mbms.artist WHERE name ILIKE $1",
        "%"+query+"%")
    if err != nil {
        log.Printf("HandleSearch: %s, %v\n", "Artist query error", err)
        return fmt.Sprintf(server.templates["search.html"], "", "")
    }
    trackList, _ := server.GenerateTrackList(trackRows)
    releaseList := server.GenerateReleaseList(releaseRows)
    artistList := server.GenerateArtistList(artistRows)
    var (
        trackListTitle, releaseListTitle, artistListTitle string
    )
    if trackList != "" {
        trackListTitle = "Tracks"
    }
    if releaseList != "" {
        releaseListTitle = "Releases"
    }
    if artistList != "" {
        artistListTitle = "Artists"
    }
	return fmt.Sprintf(server.templates["search.html"],
		query,
		releaseListTitle,
        releaseList,
        trackListTitle,
        trackList,
        artistListTitle,
        artistList)
}

func (server *Server) HandlePlaylists() string {
    rows, err := server.db.Query(server.ctx,
    "SELECT name,id FROM mbms.playlist")
    if err != nil {
        log.Printf("Function: %s\n", "Query error")
    }
    defer rows.Close()
    var id int
    var name string
    var playlists string 
	for rows.Next() {
		err := rows.Scan(&name, &id)
        if err != nil {
            log.Printf("Function: %s\n", "Query error")
        }
        playlists += fmt.Sprintf(server.templates["playlist-li.html"], id, name) 
    }
    return fmt.Sprintf(server.templates["playlists.html"], playlists)
}

func (server *Server) HandlePlaylist(id string) string {
    var name string
    err := server.db.QueryRow(server.ctx,
    "SELECT name FROM mbms.playlist WHERE id = $1", id).Scan(&name)
    if err != nil {
        log.Printf("Function: %s\n", "Query error")
    }
    rows, err := server.db.Query(server.ctx,
    "SELECT mbid,name,artist_credit,length,number FROM mbms.track WHERE mbid IN (SELECT track FROM mbms.playlist_track WHERE playlist = $1)",
    id)
    if err != nil {
        log.Printf("Function: %s\n", "Query error")
    }
    defer rows.Close()
    var (
        tracks, mbid, trackName string
        artistCredit, duration, number int
    )
    for rows.Next() {
        err := rows.Scan(&mbid, &trackName, &artistCredit, &duration, &number)
        if err != nil {
            log.Printf("Function: %s\n", "Query error")
        }
        tracks += fmt.Sprintf(server.templates["track-li.html"], 
			number,
			trackName,
			server.GenerateArtistCredit(artistCredit),
            mbid,
			ConvertMsToTime(duration))
    }
    return fmt.Sprintf(server.templates["playlist.html"], name, tracks)
}
func (server *Server) HandlePlaylistCreate(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    name := r.FormValue("name")
    if name == "" {
        log.Printf("Function: %s\n", "Empty name")
        http.Error(w, "Empty name", http.StatusBadRequest)
        return
    }
    _, err := server.db.Exec(server.ctx,
    "INSERT INTO playlist (name) VALUES ($1)", name)
    if err != nil {
        log.Printf("Function: %s\n", "Query error")
        http.Error(w, "Query error", http.StatusInternalServerError)
        return
    }
}
func (server *Server) HandlePlaylistAdd(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    playlist := r.FormValue("playlist")
    track := r.FormValue("track")
    if playlist == "" || track == "" {
        log.Printf("HandlePlaylistAdd: %s\n", "Empty playlist or track")
        http.Error(w, "Empty playlist or track", http.StatusBadRequest)
        return
    }
    _, err := server.db.Exec(server.ctx,
    "INSERT INTO playlist_track (playlist,track) VALUES ($1,$2)", playlist, track)
    if err != nil {
        log.Printf("HandlePlaylistAdd: %s\n", "Query error")
        http.Error(w, "Query error", http.StatusInternalServerError)
        return
    }
}
func (server *Server) HandlePlaylistRemove(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    playlist := r.FormValue("playlist")
    track := r.FormValue("track")
    if playlist == "" || track == "" {
        log.Printf("HandlePlaylistRemove: %s\n", "Empty playlist or track")
        http.Error(w, "Empty playlist or track", http.StatusBadRequest)
        return
    }
    _, err := server.db.Exec(server.ctx,
    "DELETE FROM mbms.playlist_track WHERE playlist = $1 AND track = $2", playlist, track)
    if err != nil {
        log.Printf("HandlePlaylistRemove: %s\n", "Query error")
        http.Error(w, "Query error", http.StatusInternalServerError)
        return
    }
}
func (server *Server) HandlePlaylistList(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    track_mbid := r.FormValue("track")
    rows, err := server.db.Query(server.ctx,
    "SELECT name,id FROM mbms.playlist")
    if err != nil {
        log.Printf("HandlePlaylistList: %s\n", "Query error")
        http.Error(w, "Query error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()
    var id int
    var name string
    var playlists string 
    for rows.Next() {
        err := rows.Scan(&name, &id)
        if err != nil {
            log.Printf("HandlePlaylistList: %s\n", "Query error")
            http.Error(w, "Query error", http.StatusInternalServerError)
            return
        }
        playlists += fmt.Sprintf(server.templates["playlist-chooser-li.html"], id, name, track_mbid) 
    }
    fmt.Fprintf(w, fmt.Sprintf(server.templates["playlists.html"], playlists))
}
func main() {
	log.Println("Started server")
	server := Server{
        wg: sync.WaitGroup{},
        cache: make(map[string]struct{}),
        templates: make(map[string]string),
        urls: make(map[string]string),
        ctx: context.Background(),
        artistCache: sync.Map{},
        artistCreditCache: sync.Map{},
        artistCreditNameCache: sync.Map{},
        bufferPool: sync.Pool{
            New: func() interface{} {
                return new(bytes.Buffer)
            },
        },
        ffprobePool: sync.Pool{
            New: func() interface{} {
                return  Ffprobe{}
            },
        },
    }
	// TODO: Find a better way instead of using a map
	server.LoadTemplate("404.html")
	server.LoadTemplate("base.html")
	server.LoadTemplate("artist.html")
	server.LoadTemplate("artist-link.html")
	server.LoadTemplate("artist-li.html")
	server.LoadTemplate("release-li.html")
	server.LoadTemplate("release.html")
	server.LoadTemplate("track-li.html")
	server.LoadTemplate("track.json")
	server.LoadTemplate("search.html")
    server.LoadTemplate("playlists.html")
    server.LoadTemplate("playlist-li.html")
    server.LoadTemplate("playlist-chooser-li.html")
    server.LoadTemplate("playlist.html")
	config, err := pgxpool.ParseConfig("postgres://localhost:5432/mbms?pool_max_conns=100")
	server.db, err = pgxpool.NewWithConfig(server.ctx, config)
	if err != nil {
		log.Fatal(err)
	}
	defer server.db.Close()
    server.batch = &pgx.Batch{}
	// clear db
	server.batch.Queue(" DROP SCHEMA IF EXISTS mbms CASCADE") 
	server.batch.Queue("CREATE SCHEMA mbms")
    // create tables
    server.batch.Queue("CREATE TABLE mbms.artist (mbid VARCHAR(36) NOT NULL, name VARCHAR, unique (mbid))")
    server.batch.Queue("CREATE TABLE mbms.artist_credit (name VARCHAR NOT NULL, id INTEGER NOT NULL, unique (name))")
    server.batch.Queue("CREATE TABLE mbms.artist_credit_name (artist_credit INTEGER NOT NULL, name VARCHAR NOT NULL, artist_mbid VARCHAR(36) NOT NULL, unique (artist_credit, artist_mbid))")
    server.batch.Queue("CREATE TABLE mbms.release (mbid VARCHAR(36) NOT NULL, name VARCHAR NOT NULL, artist_credit INTEGER NOT NULL, date VARCHAR, cover_url VARCHAR, unique (mbid))")
    server.batch.Queue("CREATE TABLE mbms.track (mbid VARCHAR(36) NOT NULL, name VARCHAR NOT NULL, artist_credit INTEGER NOT NULL, release_mbid VARCHAR(36) NOT NULL, number INTEGER NOT NULL, length INTEGER NOT NULL, url VARCHAR, unique (mbid))")
    server.batch.Queue("CREATE TABLE mbms.playlist (id SERIAL PRIMARY KEY, name VARCHAR NOT NULL, unique (name))")
    server.batch.Queue("CREATE TABLE mbms.playlist_track (playlist INTEGER NOT NULL, track VARCHAR(36) NOT NULL, unique (playlist, track))")
    server.batch.Queue("SET search_path TO mbms")
    err = server.db.SendBatch(server.ctx, server.batch).Close()
    if err != nil {
        log.Fatal(err)
    }
	log.Println("Importing library... ")
	server.wg.Add(1)
    go server.ImportLibrary("/backup/encoded")
    server.wg.Wait()
	log.Println("Importing library... done")
    log.Println("Sending batch... ")
    err = server.db.SendBatch(server.ctx, server.batch).Close()
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Sending batch... done")
	router := httprouter.New()
	// There has to be a simpler way to do this
	router.GET("/artist/:mbid", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		server.Handle(w, r, ps, server.HandleArtist, func(str string) string {
			return fmt.Sprintf(server.templates["base.html"], str)
		})
	})
	router.GET("/release/:mbid", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		server.Handle(w, r, ps, server.HandleRelease, func(str string) string {
			return fmt.Sprintf(server.templates["base.html"], str)
		})
	})
	router.GET("/track/:mbid", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		server.Handle(w, r, ps, server.HandleTrack, func(str string) string {
			return str
		})
	})
	router.GET("/ajax/artist/:mbid", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		server.Handle(w, r, ps, server.HandleArtist, func(str string) string {
			return str
		})
	})
	router.GET("/ajax/release/:mbid", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		server.Handle(w, r, ps, server.HandleRelease, func(str string) string {
			return str
		})
	})
	router.GET("/search", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w, fmt.Sprintf(server.templates["base.html"], server.HandleSearch(r.URL.Query().Get("q"))))
	})
	router.GET("/ajax/search", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w, server.HandleSearch(r.URL.Query().Get("q")))
	})
	router.GET("/playlists", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w, fmt.Sprintf(server.templates["base.html"], server.HandlePlaylists()))
	})
	router.GET("/ajax/playlists", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		fmt.Fprintf(w, server.HandlePlaylists())
	})
    router.GET("/playlist/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	    id := ps.ByName("id")
		fmt.Fprintf(w, fmt.Sprintf(server.templates["base.html"], server.HandlePlaylist(id)))
	})
    router.GET("/ajax/playlist/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	    id := ps.ByName("id")
		fmt.Fprintf(w, server.HandlePlaylist(id))
	})
    router.POST("/rest/playlist/create", server.HandlePlaylistCreate)
    router.POST("/rest/playlist/add", server.HandlePlaylistAdd)
    router.POST("/rest/playlist/remove", server.HandlePlaylistRemove)
    router.POST("/rest/playlist/list", server.HandlePlaylistList)
	router.GET("/media/:mbid", server.HandleMedia)
	router.ServeFiles("/static/*filepath", http.Dir("static"))
	log.Fatal(http.ListenAndServe(":8080", router))
}
