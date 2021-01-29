elem("refresh").addEventListener("click", search)

function search(ev) {
    ev.preventDefault()
    const name = elem("name").value
    const school = elem("school").value
    const year = elem("month").value
    const month = elem("month").value
    const subject = elem("subject").value
    const theme = elem("theme").value

    fetch(`/search?name=${name}&school=${school}&year=${year}&subject=${subject}&theme=${theme}&month=${month}`).then(re => {
        re.json().then( js => {
            console.log(js)
        })    
    })
}