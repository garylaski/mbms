let playlist = [];
let current_track = 0;
function add_track(id) {
    let track = get_track_json(id);
    playlist.push(track);
}
function toggle() {
    audio = document.getElementById("audio");
    if (audio.paused) {
        audio.play();
        document.getElementById("play").id = "pause";
    } else {
        audio.pause();
        document.getElementById("pause").id = "play";
    }
}
function get_track_json(id) {
    let url = "/rest/getTrack?id=" + id;
    let request = new XMLHttpRequest();
    request.open("GET", url, false);
    request.send(null);
    return JSON.parse(request.responseText);
}
function play_track(id) {
    let track = get_track_json(id);
    playlist = [track];
    current_track = 0;
    let audio = document.getElementById("audio");
    //replace div with play id to pause
    if (document.getElementById("play")) {
        document.getElementById("play").id = "pause";
    }
    audio.src = "/media/" + playlist[current_track].url;
    set_player(track);
    audio.play();
}
function play_tracks(ids) {
    playlist = [];
    add_tracks(ids);
    current_track = 0;
    let audio = document.getElementById("audio");
    if (document.getElementById("play")) {
        document.getElementById("play").id = "pause";
    }
    audio.src = "/media/" + playlist[current_track].url;
    set_player(playlist[current_track]);
    audio.play();
}
function add_tracks(ids) {
    // Add tracks to playlist
    for (i in ids) {
        add_track(ids[i]);
    }
}
function set_player(track) {
    document.getElementById("player-name").innerHTML = track.name;
    document.getElementById("player-length").innerHTML = new Date(track.length).toISOString().slice(14, 19);
    document.getElementById("player-artist-credit").innerHTML = track.artist_credit_html;
    document.getElementById("track-cover").src = "/media/" + track.cover_url;
    document.getElementById("release-url").href = "/release/" + track.release_mbid;
    document.getElementById("release-url").onclick = function() { ajax("/release/" + track.release_mbid); return false;};
    document.getElementById("seek-slider").max = track.length;
}
function next_track() {
    current_track++;
    if (current_track >= playlist.length) {
        // Stop playing
        return;
    }
    let audio = document.getElementById("audio");
    audio.src = "/media/" + playlist[current_track].url;
    set_player(playlist[current_track]);
    audio.play();
}
function prev_track() {
    current_track--;
    if (current_track < 0) {
        // Stop playing
        return;
    }
    let audio = document.getElementById("audio");
    audio.src = "/media/" + playlist[current_track].url;
    set_player(playlist[current_track]);
    audio.play();
}

function ajax_load(url) {
    let xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) {
            document.getElementById("content").innerHTML = xhr.responseText;
            document.title = document.getElementById("title").innerHTML;
            localStorage.setItem("last_url", url);
        }
    };
    xhr.open("GET", "/ajax/" + url, true);
    xhr.send();
}

function ajax(url) {
    ajax_load(url);
    window.history.pushState({}, "", url);
}

function update_player_time() {
    let audio = document.getElementById("audio");
    let player_current_time = document.getElementById("player-current-length");
    let seek_slider = document.getElementById("seek-slider");
    current_time = new Date(audio.currentTime*1000).toISOString().slice(14, 19);
    player_current_time.innerHTML = current_time;
    seek_slider.value = audio.currentTime*1000;
}

// After document loads add event listener to audio element
document.addEventListener("DOMContentLoaded", function() {
    document.getElementById("audio").addEventListener("ended", next_track);
});

window.onpopstate = function (e) {
    ajax_load(window.location.pathname);
}
window.onload = function() {
    let audio = document.getElementById("audio");
    audio.addEventListener("timeupdate", update_player_time);
}

