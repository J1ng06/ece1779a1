document.getElementById('sendBtn').addEventListener('click', handleFileSend, false);
document.getElementById('image').addEventListener('change', handleFileSelect, false);
document.getElementById('resetBtn').addEventListener('click', handleReset, false);

var imagecounts;
var pageSize = 8;

var modal = document.getElementById('modal');
window.onclick = function (event) {
    if (event.target == modal) {
        modal.style.display = "none";
    }
}

function handleReset() {
    document.getElementById('preview').innerHTML = "";
}

function handleFileSend(evt) {

    var tag = document.getElementById('image'); // FileList object

    // Loop through the FileList and render image files as thumbnails.
    for (var i = 0, f; f = tag.files[i]; i++) {

        // Only process image files.
        if (!f.type.match('image.*')) {
            continue;
        }

        var reader = new FileReader();

        // Closure to capture the file information.
        reader.onload = (function (theFile) {
            return function (e) {

                var xhr = new XMLHttpRequest(),
                    path = "image/upload?username=" + window.location.href.split('username=')[1];

                data = JSON.stringify({
                    name: theFile.name,
                    image: e.target.result
                });

                xhr.onreadystatechange = function () {
                    if (xhr.status == 200 && xhr.readyState == 4) {
                        window.location.replace(xhr.getResponseHeader("Location").toLowerCase())
                    }
                }
                xhr.open("POST", path, true);
                xhr.send(data);

            };
        })(f);

        // Read in the image file as a data URL.
        reader.readAsDataURL(f);
    }
}

function handleFileSelect(evt) {
    var files = evt.target.files; // FileList object

    // Loop through the FileList and render image files as thumbnails.
    for (var i = 0, f; f = files[i]; i++) {

        // Only process image files.
        if (!f.type.match('image.*')) {
            continue;
        }

        var reader = new FileReader();

        // Closure to capture the file information.
        reader.onload = (function (theFile) {
            return function (e) {
                var span = document.createElement('span');
                span.innerHTML = ['<img class="thumb" src="', e.target.result,
                    '" title="', escape(theFile.name), ' width="200px" height="200px"/>'].join('');
                document.getElementById('preview').insertBefore(span, null);
            };
        })(f);

        // Read in the image file as a data URL.
        reader.readAsDataURL(f);
    }
}

function initalLoad() {
    var xhr = new XMLHttpRequest();

    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {
            var imagecounts = JSON.parse(xhr.response);
            console.log(imagecounts);

            if (imagecounts > pageSize) {
                var pagination = document.getElementById("pagination");
                for (var i = 0; i < imagecounts; i = i + pageSize) {

                    var a = document.createElement("a")
                    a.setAttribute("id", "page" + (i / pageSize + 1));
                    a.setAttribute("onclick", ['loadThumbnails(', i / pageSize, ', this.id)'].join(""));
                    a.innerHTML = i / pageSize + 1;
                    if (i === 0) {
                        a.setAttribute("class", "active");
                    } else {
                        a.setAttribute("class", "inactive");
                    }
                    pagination.insertBefore(a, null)
                }

            }
            var page = 0;
            loadThumbnails(page)
        }
    }
    xhr.open("Get", "image/userimagecount?username=" + window.location.href.split('username=')[1], true);
    xhr.send();
}

function loadThumbnails(page, id) {
    if (id) {
        var pages = [].slice.call(document.getElementById("pagination").children);
        pages.forEach(function (e) {
            e.setAttribute("class", "inactive");
            if (e.id === id) {
                e.setAttribute("class", "active");
            }
        });
    }
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function () {
        if (xhr.readyState == 4 && xhr.status == 200) {
            var response = JSON.parse(xhr.response);
            console.log(response)
            document.getElementById('column').innerHTML = "";
            var thumbnails = response.map(i => i.thumbnail);

            thumbnails.forEach(function (location) {
                loadThumbnail(location)
            });
        }
    }
    xhr.open("Get", "image/userimages?username=" + window.location.href.split('username=')[1] + "&page=" + page, true);
    xhr.send();
}

function loadThumbnail(location) {
    var xhr = new XMLHttpRequest();
    xhr.onreadystatechange = function () {
        if (xhr.readyState == 4 && xhr.status == 200) {
            var response = xhr.response;
            var img = document.createElement('img');
            img.className = "thumb";
            img.src = response;
            img.style = "width:25%";
            img.setAttribute("id", location)
            img.setAttribute("onclick", "showModal(this.id)");
            document.getElementById('column').insertBefore(img, null);
        }
    }

    xhr.open("Get", "image/getimage?username=" + window.location.href.split('username=')[1] + "&location=" + location, true);
    xhr.send();
}

function showModal(id) {

    var span = document.getElementsByClassName("close")[0];
    var images = document.getElementById("images");
    images.innerHTML = "";
    ["original", "t1", "t2", "t3"].forEach(function (e) {
        loadImage(id.replace("thumbnail", e))
    });

    modal.style.display = "block";
    span.onclick = function () {
        modal.style.display = "none";
    }


}

function loadImage(location) {
    var xhr = new XMLHttpRequest();
    var images = document.getElementById("images");
    xhr.onreadystatechange = function () {
        if (xhr.readyState == 4 && xhr.status == 200) {
            var response = xhr.response;
            var img = document.createElement('img');
            img.src = response;
            img.id = location;
            img.style = "width:100%";
            images.appendChild(img);
        }
    }

    xhr.open("Get", "image/getimage?username=" + window.location.href.split('username=')[1] + "&location=" + location, true);
    xhr.send();
}

function logout() {
    var xhr = new XMLHttpRequest();
    var data = {
        "username": window.location.href.split('username=')[1],
        "password": "logout"
    }
    xhr.onreadystatechange = function () {
        if (xhr.status == 200 && xhr.readyState == 4) {
            window.location.replace(xhr.getResponseHeader("Location").toLowerCase())
        }
    }

    xhr.open("Post", "user/logout", true);
    xhr.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
    console.log(JSON.stringify(data));
    xhr.send(JSON.stringify(data));
}