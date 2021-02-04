const id = new URLSearchParams(window.location.search).get("id")
const error = elem("error")
const info = elem("info")
const content = elem("content")
const comments = elem("comments")

const commentA = elem("comment-a")
const add = elem("add")
const like = elem("like")

loadAccount()

request("publicnote", {id: id}).then(j => {
    const err = getErr(j)
    if(err) {
        error.innerHTML = err
        return
    }


    const n = j.Note
    searchParams.filter(e => !["school", "author"].includes(e)).forEach(e => {
        const key = e[0].toUpperCase() + e.substring(1)
        info.appendChild(infElem(e, n[key]))
    }) 

    info.appendChild(infElem("school", schools[n.School]))

    request("publicaccount", {id: n.Author}).then(j => {
        var err = getErr(j) 
        if (err == undefined) {
           err = `<a href="account.html?id=${n.Author}" style="color:wheat;">${j.Account.Name}</a>`
        }
        info.appendChild(infElem("author", err))

        content.innerHTML = new Markdown(j.Account.Cfg.Colors).convert(n.Content)
    })
})


// TODO implement size limits for comments in backend
add.onclick = function(e) {
    e.preventDefault()

    error.innerHTML = ""

    request("comment", {id: id, target: "note"}, {
        method: "POST",
        headers: {
            'content-type': 'text/plain'
        },
        body: commentA.value,
    }).then(j => {
        const err = getErr(j)
        if (err) {
            error.innerHTML = err
            return
        }


    })
}

like.onclick = function(e) {
    e.preventDefault()
    error.innerHTML = ""

    request("like", {id: id, target: "note"}).then(j => {
        const err = getErr(j)
        if(err){
            error.innerHTML = err
            return
        }

        if(j.Like) {
            like.innerHTML = "dislike"
        } else {
            like.innerHTML = "like"
        }
    })
}

function comment(user, likes, content) {
    const img = document.createElement("img")
    
}

function addLikeEvent(id) {
    const icon = elem(`like${id}`)
    const counter = elem(`counter${id}`)

    like.onclick = function(e) {
        icon.setAttribute("src", "")
    }
}

addLikeEvent("")

function infElem(key, value) {
    const div = document.createElement("div")
    div.classList.add("inf")
    const span = document.createElement("span")
    span.classList.add("bold")
    span.innerHTML = key + ": "
    div.appendChild(span)
    div.innerHTML += value
    return div
}