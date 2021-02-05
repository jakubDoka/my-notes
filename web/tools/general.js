class UndoState {
    constructor(text, start, end) {
        this.text = text
        this.start = start
        this.end = end
    }
}

function getCookie(cname) {
    var name = cname + "=";
    var decodedCookie = decodeURIComponent(document.cookie);
    var ca = decodedCookie.split(';');
    for(var i = 0; i <ca.length; i++) {
      var c = ca[i];
      while (c.charAt(0) == ' ') {
        c = c.substring(1);
      }
      if (c.indexOf(name) == 0) {
        return c.substring(name.length, c.length);
      }
    }
    return "";
}

function setCookie(cname, cvalue, exdays) {
    var d = new Date();
    d.setTime(d.getTime() + (exdays*24*60*60*1000));
    var expires = "expires="+ d.toUTCString();
    document.cookie = cname + "=" + cvalue + ";" + expires + ";path=/";
}

function elem(id) {
    return document.getElementById(id)
}

function elemByClass(name) {
    return document.getElementsByClassName(name)
}

function insert(idx, str, sub) {
    return str.substring(0, idx) + sub + str.substring(idx)
}

function replace(start, end, str, sub) {
    return str.substring(0, start) + sub + str.substring(end)
}

function assertLogin(message) {
    const dat = getCookie("user")
    if(dat == ""){
        window.location.href = "login.html"
        window.alert(message)
    }
    return dat
}

function gotoLogin(message) {
    window.location.href = "login.html"
    if(message != "") {
        window.alert(message)
    }
}

function IsFilled(...params) {
    for(var i in params) {
        if(params[i].value == "") {
            return false
        }
    }

    return true
}

async function sha256(message) {
    // encode as UTF-8
    const msgBuffer = new TextEncoder().encode(message);                    

    // hash the message
    const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);

    // convert ArrayBuffer to Array
    const hashArray = Array.from(new Uint8Array(hashBuffer));

    // convert bytes to hex string                  
    const hashHex = hashArray.map(b => ('00' + b.toString(16)).slice(-2)).join('');
    return hashHex;
}

const searchParams = ["name", "school", "year", "month", "subject", "theme", "author"]
const schools = ["none", "elementary-middle", "high", "university"]

async function search(query) {
    return await fetch(buildRequest(searchParams, "search", query)).then(handleResponse)
}

function searchSetup() {
    return {
        name: elem("name"),
        school: elem("school"),
        year: elem("year"),
        month: elem("month"),
        subject: elem("subject"),
        theme: elem("theme"),
        author: elem("author"),
    }
}

async function request(command, params, init) {
    if(params == undefined) {
        return await fetch(command, init).then(re => re.json())
    }

    return await fetch(buildRequest(Object.keys(params), command, params), init).then(handleResponse)
}

async function handleResponse(re) {
    if(re.status != 200) {
        console.error(await re.text())
        return
    }
    return await re.json()
}

function buildRequest(params, name, provided) {
    var url = "/" + name + "?"
    params.forEach((e, i) => {
        url += e + "="
        if(provided[e] != undefined) {
            if(provided[e].value != undefined) {
                url += provided[e].value
            } else {
                url += provided[e]
            }
        }
        if(i != params.length-1) {
            url += "&"
        }
    })
    return url
}

const invalidNameMessage = "name cannot contain spaces nor start with #"

function IsValidName(name) {
    return !name.includes(" ") && !name.startsWith("#")
}

var hexDigits = new Array("0","1","2","3","4","5","6","7","8","9","a","b","c","d","e","f"); 

//Function to convert rgb color to hex format
function rgb2hex(rgb) {
    rgb = rgb.match(/^rgb\((\d+),\s*(\d+),\s*(\d+)\)$/);
    return "#" + hex(rgb[1]) + hex(rgb[2]) + hex(rgb[3]);
}

function hex(x) {
    return isNaN(x) ? "00" : hexDigits[(x - x % 16) / 16] + hexDigits[x % 16];
}

function getErr(response) {
    if((response.Status) ? response.Status == "success" : response.Resp.Status == "success") {
        return undefined
    } else {
        const err = (response.Status) ? response.Status : response.Resp.Status
        console.trace(err)
        return err
    }
}

class Time {
    constructor(milliseconds){
        this.mil = milliseconds
        this.sec = this.mil / 1000
        this.min = this.sec / 60
        this.hour = this.min / 60
        this.day = this.hour / 24
        this.year = this.day / 365

        for(var pr in this) {
            this[pr] = Math.round(this[pr])
        }
    }

    toString() {
        var res = "nothing"
        for(var pr in this) {
            if(this[pr] != 0){
                res = `${this[pr]}${pr}s`
            }
        }
        return res
    }
}

const defaultColors = ["#b03830", "#b0972a", "#5d62f0"] 

var user = undefined

async function loadAccount(addition) {
    const j = await request("account")
    const err = getErr(j)
    if (err) {
        gotoLogin(err)
        return
    } else {
        user = j.Account
        if(user.Cfg.Colors == null) {
            user.Cfg.Colors = defaultColors
        }
    }

    if (addition) {
        addition()
    }
}

async function loadText(path) {
    return await fetch(path).then(r => r.text())
}

async function loadTemplate(path, args) {
    const template = await fetch(path).then(r => r.text())
    return Format(template, args)
}

function format(str, args) {
    if (args instanceof Array) {
        args.forEach((e, i) => {
            str = str.replaceAll(`{${i}}`, e)
        })
    } else {
        for(var pr in args) {
            str = str.replaceAll(`{${pr}}`, args[pr])
        }
    }

    return str
}

function registerLikeButton(like, counter, data, target, id) {
    setLikeState(like, counter, data.State, data.Count)
    console.log("pp")
    like.onclick = async function(e) {
        e.preventDefault()
        console.log("pp")
        request("like", {id: id, target: target, change: true}).then(j => {
            const err = getErr(j)
            if (err) {
                alert(err)
                return
            }
            setLikeState(like, counter, j.State, j.Count)
        }) 
    }
}

function setLikeState(like, counter, state, count) {
    like.setAttribute("src", state ? "assets/liked.png": "assets/not-liked.png")
    counter.innerHTML = count.toString()
}

async function embed(path, target) {
    const e = document.getElementById(target)
    const text = await loadText(path)
    e.innerHTML = text
    Rerun(e)
}

function Rerun(elem) {
    elem.childNodes.forEach(element => {
        if(element.tagName == "SCRIPT") {
            var data = (element.text || element.textContent || element.innerHTML || "" ),
            head = document.getElementsByTagName("head")[0] ||
                      document.documentElement,
            script = document.createElement("script");
    
            script.type = "text/javascript";
            try {
            // doesn't work on ie...
            script.appendChild(document.createTextNode(data));      
            } catch(e) {
            // IE has funky script nodes
            script.text = data;
            }
        
            head.insertBefore(script, head.firstChild);
            head.removeChild(script);
        }
        Rerun(element)
    })
}

embed("components/menu.html", "menu")
