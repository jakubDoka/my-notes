const refresh = elem("refresh")

const results = elem("results")

const query = searchSetup()

refresh.onclick = function(ev) {
    ev.preventDefault()
    search(query).then(j => {
        const err = getErr(j)
        if(err) {

        } else {
            for(var idx in j.Results) {
                const i = idx
                const res = j.Results[i]
                const elapsed = new Date().getTime() - res.BornDate
                results.innerHTML += `
                <div class="stats">
                    <a href="view.html?id=${res.ID}" class="title">${res.Name}</a>
                    <img src="assets/like.png">
                    <span class="title">${res.Likes}</span>
                    <span class="date">created ${new Time(elapsed).toString()} ago</span> 
                </div>
                <div class="text-box preview" id="preview${i}"></div>
                `
                
                request("config", {id: res.Author}).then(j => {
                    const err = getErr(j)
                    const preview = elem(`preview${i}`)
                    console.log(j, res.Content)
                    if (err) {
                        preview.innerHTML = err
                    } else {
                        const m = new Markdown((j.Cfg.Colors == null) ? defaultColors : j.Cfg.Colors)
                        preview.innerHTML = m.convert(res.Content)
                        console.log(m.convert("asdasd"))
                    }
                })
            }
        }
    })
}

refresh.click()