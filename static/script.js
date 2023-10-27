var audio, play_button, key_buffer, num_buffer, index, playlist, current_track, volume;
let prev_volume = 100;
function add_track(id) {
    playlist.push(get_track_json(id));
}
function get_track_json(mbid) {
    let request = new XMLHttpRequest();
    request.open("GET", "/track/" + mbid, false);
    request.send(null);
    return JSON.parse(request.responseText);
}
function play_track(mbid) {
    audio.src = "/media/" + mbid;
    play_button.classList = "pause";
    audio.play();
    let track = get_track_json(mbid);
    playlist = [track];
    current_track = 0;
    set_player(track);
}
function play_tracks(mbids) {
    playlist = [];
    add_tracks(mbids);
    current_track = 0;
    audio.src = "/media/" + playlist[current_track].mbid;
    audio.play();
    play_button.classList = "pause";
    set_player(playlist[current_track]);
}
function add_tracks(ids) {
    for (var i in ids) {
        add_track(ids[i]);
    }
}
function set_player(track) {
    let player = document.getElementById("player");
    player.querySelector(".track-name").innerHTML = track.name;
    player.querySelector(".cover").src = "/media/" + track.release_mbid;
    player.querySelector("#player-length").innerHTML = new Date(track.length).toISOString().slice(14, 19);
    player.querySelector(".release-url").href = "/release/" + track.release_mbid;
    player.querySelector(".release-url").onclick = function() { ajax("/release/" + track.release_mbid); return false;};
    player.querySelector("#seek-slider").max = track.length;
    player.querySelector(".artist-credit").innerHTML = track.artist_credit;
}
function next_track() {
    current_track++;
    if (current_track < playlist.length) {
        audio.src = "/media/" + playlist[current_track].mbid;
        set_player(playlist[current_track]);
        audio.play();
    }
}
function prev_track() {
    current_track--;
    if (current_track >= 0) {
        audio.src = "/media/" + playlist[current_track].mbid;
        set_player(playlist[current_track]);
        audio.play();
    }
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
function rest(url, params, ajax_url) {
    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/rest/" + url, false);
    xhr.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
    xhr.send(new URLSearchParams(params).toString());
    if (ajax_url) {
        ajax_load(ajax_url);
    }
    return xhr.responseText;
}
function update_player_time() {
    document.getElementById("player-current-length").innerHTML = new Date(audio.currentTime*1000).toISOString().slice(14, 19);
    document.getElementById("seek-slider").value = audio.currentTime*1000;
}
function toggle_vol(e) {
    let slider = document.querySelector(".volume-slider");
    if (e.classList == "volume") {
        e.classList = "muted";
        prev_volume = slider.value;
        slider.value = 0;
        change_volume(0);
    } else {
        e.classList = "volume";
        slider.value = prev_volume;
        change_volume(prev_volume);
    }
}
function seek(value) {
    audio.currentTime = value/1000;
}
function change_volume(value) {
    audio.volume = value/100;
    setCookie("volume", value, 365);
}
window.onpopstate = function (e) {
    ajax_load(window.location.pathname + window.location.search);
}
window.onload = function() {
    play_button = document.querySelector("#play_button");
    volume = document.querySelector(".volume-slider");
    volume.value = getCookie("volume");
    audio = document.getElementById("audio");
    audio.volume = volume.value/100;
    audio.addEventListener("timeupdate", update_player_time);
    audio.addEventListener("ended", next_track);
}
function setCookie(c_name,value,exdays)
{
    var exdate=new Date();
    exdate.setDate(exdate.getDate() + exdays);
    var c_value=escape(value) + ((exdays==null) ? "" : ("; expires="+exdate.toUTCString()));
    document.cookie=c_name + "=" + c_value;
}
function getCookie(c_name)
{
    var i,x,y,ARRcookies=document.cookie.split(";");
    for (i=0; i<ARRcookies.length; i++)
    {
        x=ARRcookies[i].substr(0,ARRcookies[i].indexOf("="));
        y=ARRcookies[i].substr(ARRcookies[i].indexOf("=")+1);
        x=x.replace(/^\s+|\s+$/g,"");
        if (x==c_name)
        {
            return unescape(y);
        }
    }
}
function toggle_play() {
    if (audio.src) {
        if (audio.paused) {
            play_button.classList = "pause";
            audio.play();
        } else {
            play_button.classList = "play";
            audio.pause();
        }
    }
}
function open_playlist_chooser(mbid) {
    let inner_html = rest("playlist/list", {"track":mbid}, false);
    let chooser = document.getElementById("chooser");
    chooser.style.display = "block";
    chooser.innerHTML = inner_html;
}
function add_to_playlist(playlist_id, mbid) {
    rest("playlist/add", {"playlist":playlist_id, "track":mbid}, false);
    document.getElementById("chooser").style.display = "none";
}
