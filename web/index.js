const refresh = elem("refresh")
const query = searchSetup()

refresh.onclick = function(ev) {
    ev.preventDefault()
    search(query).then(j => {
        if(j.Resp.Status == "success")
    })
}