let playlist = [];
let current_track = 0;
function add_track(mbid) {
    playlist.push(mbidToJSONTrack(mbid));
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
function mbidToJSONTrack(mbid) {
    let url = "../../track/" + mbid;
    let request = new XMLHttpRequest();
    request.open("GET", url, false);
    request.send(null);
    return JSON.parse(request.responseText);
}
function play_track(mbid) {
    let track = mbidToJSONTrack(mbid);
    playlist = [track];
    current_track = 0;
    let audio = document.getElementById("audio");
    //replace div with play id to pause
    if (document.getElementById("play")) {
        document.getElementById("play").id = "pause";
    }
    audio.src = "../../media/" + playlist[current_track].url;
    set_player(track);
    audio.play();
}
function play_tracks(mbids) {
    playlist = [];
    add_tracks(mbids);
    current_track = 0;
    let audio = document.getElementById("audio");
    if (document.getElementById("play")) {
        document.getElementById("play").id = "pause";
    }
    audio.src = "../../media/" + playlist[current_track].url;
    set_player(playlist[current_track]);
    audio.play();
}
function add_tracks(mbids) {
    // Add tracks to playlist
    for (i in mbids) {
        add_track(mbids[i]);
    }
}
function set_player(track) {
    document.getElementById("player-name").innerHTML = track.name;
    document.getElementById("player-length").innerHTML = new Date(track.length).toISOString().slice(14, 19);
    document.getElementById("player-artist-credit").innerHTML = track.artistCredit;
    document.getElementById("track-cover").src = "/media/" + track.releaseCoverUrl;
    document.getElementById("seek-slider").max = track.length;
}
function next_track() {
    current_track++;
    if (current_track >= playlist.length) {
        // Stop playing
        return;
    }
    let audio = document.getElementById("audio");
    audio.src = "../../media/" + playlist[current_track].url;
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
    audio.src = "../../media/" + playlist[current_track].url;
    set_player(playlist[current_track]);
    audio.play();
}

function ajax_load(url) {
    let xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function() {
        if (xhr.readyState == 4 && xhr.status == 200) {
            let html = document.createElement("html");
            html.innerHTML = xhr.responseText;
            localStorage.setItem('response', xhr.responseText);
            document.getElementById("content").innerHTML = html.querySelector("#content").innerHTML;
        }
    };
    xhr.open("GET", url, true);
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
    ajax(window.location.href);
}
window.onload = function() {
    let audio = document.getElementById("audio");
    audio.addEventListener("timeupdate", update_player_time);
}

