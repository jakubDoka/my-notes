const userData = assertLogin("you have to login to access your account (surprise indeed)")


const nm = elem("name")

const nameA = elem("name-a")
const saveB = elem("save-b")
const backB = elem("back-b")
const editB = elem("edit-b")
const hint = elem("hint")

const colorP = elem("color-pick")
const addColorB = elem("add-color")
const colorsSelect = elem("colors-select")
const trash = elem("trash")
var counter = 0

const drafts = elem("drafts")
const published = elem("published")

const error = elem("error")

trash.ondrop = function(e) {
    console.log("ok")
    elem(e.dataTransfer.getData("id")).remove();
}

trash.ondragover = function(ev) {
    ev.preventDefault()
}


addColorB.onclick = function(e) {
    e.preventDefault()
    createColor(colorP.value)
}

function createColor(color) {
    var e = document.createElement("div")
    e.classList.add("color-display")
    e.id = counter
    e.draggable = true
    e.style.backgroundColor = color

    e.ondragstart = function(ev) {
        ev.dataTransfer.setData("id", e.id)
    }
    e.ondragover = function(ev) {
        ev.preventDefault()
    }
    e.ondrop = function(ev) {
        colorsSelect.insertBefore(elem(ev.dataTransfer.getData("id")), e)
    }

    colorsSelect.appendChild(e)
    counter++
}

loadAccount(() => {
    loadProfile(user.Name, user.Cfg.Colors)

    request("usernotes", {id: user.ID}).then(j => {
        const err = getErr(j)
        if(err) {
            error2.innerHTML = err
            return
        }
        
        loadText("components/draft.html").then(t => {
            for(var i in j.Drafts) {
                const n = j.Drafts[i] 
                const str = format(t, {
                    name: n.Name, 
                    subject: n.Subject, 
                    color: user.Cfg.Colors[1], 
                    theme: n.Theme,
                    year: n.Year,
                    month: n.Month,
                    id: n.ID,
                })
        
                if(n.Published) {
                    published.innerHTML += str
                } else {
                    drafts.innerHTML += str
                }
            }
        })
    })
})

editB.onclick = function(e) {
    nm.hidden = true
    editB.hidden = true

    nameA.hidden = false
    backB.hidden = false
    saveB.hidden = false
    colorP.hidden = false
    addColorB.hidden = false
    trash.hidden = false
    hint.hidden = false

}

backB.onclick = function(e) {
    nm.hidden = false
    editB.hidden = false

    nameA.hidden = true
    backB.hidden = true
    saveB.hidden = true
    colorP.hidden = true
    addColorB.hidden = true
    trash.hidden = true
    hint.hidden = true
}

saveB.onclick = function(e) {
    if(nameA.value == "") {
        error.innerHTML = "You cannot have blanc name."
        return
    }

    if(!IsValidName(nameA.value)) {
        error.innerHTML = invalidNameMessage
        return
    }

    if(colorsSelect.childNodes.length < 3) {
        error.innerHTML = "You need to have at least 3 colors selected"
        return
    }

    error.innerHTML = ""

    var colors = ""
    colorsSelect.childNodes.forEach(e => {
        colors += rgb2hex(e.style.backgroundColor) + " "
    })

    request("configure", {name: nameA.value, colors: colors.replaceAll("#", "")}).then(j => {
        const err = getErr(j)
        if(err){
            error.innerHTML = err
        } else {
            backB.click()
            nm.innerHTML = nameA.value
        }
    })
}

function loadProfile(aName, aColors) {
    nm.innerHTML = aName
    nameA.value = aName
    colorsSelect.innerHTML = ""
    aColors.forEach(e => {
        createColor(e)
    })
}