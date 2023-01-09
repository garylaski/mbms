let playlist = [];
let current_track = 0;
function add_track(url) {
    playlist.push(url);
}
function play_track(url) {
    playlist = [url];
    current_track = 0;
    let audio = document.getElementById("audio");
    audio.src = playlist[current_track];
    audio.play();
}
function next_track() {
    current_track++;
    if (current_track >= playlist.length) {
        // Stop playing
        return;
    }
    let audio = document.getElementById("audio");
    audio.src = playlist[current_track];
    audio.play();
}
function prev_track() {
    current_track--;
    if (current_track < 0) {
        // Stop playing
        return;
    }
    let audio = document.getElementById("audio");
    audio.src = playlist[current_track];
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


// After document loads add event listener to audio element
document.addEventListener("DOMContentLoaded", function() {
    document.getElementById("audio").addEventListener("ended", next_track);
});

window.onpopstate = function (e) {
    ajax(window.location.href);
}



