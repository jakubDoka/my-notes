const refresh = elem("refresh")

const results = elem("results")

const error = elem("error")

var query = undefined

loadAccount()

refresh.onclick = async function(ev) {
    ev.preventDefault()
    results.innerHTML = ""

    const j = await search(query)
    const err = getErr(j)
    if(err) {
        error.innerHTML = err
        return
    }

    const t = await loadText("components/preview.html")
    for(var idx in j.Results) {
        const i = idx
        const res = j.Results[i]
        const elapsed = new Date().getTime() - res.BornDate

        const cfg = await request("config", {id: res.Author})

        results.innerHTML += format(t, {
            id: res.ID,
            name: res.Name,
            idx: i,
            time: new Time(elapsed).toString(),
            idx: i,
            content: getErr(cfg) || new Markdown(cfg.Cfg.Colors).convert(res.Content+"..."),
        })
        
        const likeData = await request("like", {id: res.ID, target:"note", change: false})

        registerLikeButton(elem("like"+i), elem("counter"+i), likeData, "note", res.ID)
    }       
}

embed("components/school.html", "schools").then(() => {
    query = searchSetup()
    refresh.click()
})
