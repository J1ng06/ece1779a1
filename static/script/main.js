function setCookie(cname, cvalue, exdays) {
    var d = new Date();
    d.setTime(d.getTime() + exdays * 24 * 60 * 60 * 1000)
    var expires = "expires=" + d.toUTCString();
    document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}

function getCookie(cname) {
    var name = cname + "=";
    var ca = document.cookie.split(';');
    for (var i = 0; i < ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1)
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
        }
    }
    return null;
}

function checkCookie(name) {
    var user = getCookie(name)
    if (user !== "") {
        alert("Welcome again " + user);
    } else {
        user = prompt("Please enter your name:", "");
        if (user !== "" && user != null) {
            setCookie("username", user, 365);
        }
    }
}

function login() {

    var xhr = new XMLHttpRequest();

    var form = document.forms["loginForm"];
    var data = {};
    for (var i = 0, ii = form.length; i < ii; i++) {
        var input = form[i];
        if (input.name) {
            data[input.name] = input.value;
        }
    }

    xhr.onreadystatechange = function () {
        if (xhr.readyState === 4 && xhr.status === 200) {

            var response = JSON.parse(xhr.response);
            setCookie(response.name, response.value, response.lifetime);
            var location = xhr.getResponseHeader("Location").toLowerCase();
            window.location.replace(location);

        } else if (xhr.readyState === 4 && xhr.status === 404) {

            form.reset();
            alert("Username or password is not correct!");
            var location = xhr.getResponseHeader("Location").toLowerCase();
            window.location.replace(location);

        }
    }

    // send credentials
    xhr.open(form.method, "user/login", true);
    xhr.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
    console.log(JSON.stringify(data));

    xhr.send(JSON.stringify(data));
}

function register() {

    var xhr = new XMLHttpRequest();

    var form = document.forms["registerForm"];
    if (!form["username"].value) {
        form.reset();
        alert("Please enter valid username!")
        return
    }
    if (!form["password"].value) {
        form.reset();
        alert("Please enter valid password!")
        return
    }
    if (!form["passwordConfirm"].value) {
        form.reset();
        alert("Please enter valid password!")
        return
    }
    if (form["passwordConfirm"].value != form["password"].value) {
        form.reset();
        alert("Two passwords are not the same!")
        return
    }
    var data = {};
    for (var i = 0, ii = form.length; i < ii; i++) {
        var input = form[i];
        if (input.name) {
            data[input.name] = input.value;
        }
    }

    xhr.onreadystatechange = function () {
        if (xhr.readyState == 4 && xhr.status == 200) {
            var location = xhr.getResponseHeader("Location").toLowerCase();
            window.location.replace(location);
        }
        if (xhr.readyState == 4 && xhr.status == 400) {
            form.reset();
            alert("Username" + form["username"].value + " has been taken!")
        }
    }

    // send credentials
    xhr.open(form.method, "user/register", true);
    xhr.setRequestHeader("Content-Type", "application/json; charset=UTF-8");
    console.log(JSON.stringify(data));

    xhr.send(JSON.stringify(data));
}