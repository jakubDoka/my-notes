const userData = assertLogin("you have to login to use editor")

const raw = elem("raw")
const rawB = elem("raw-b")
const preview = elem("preview")
const previewB = elem("preview-b")
const shortcuts = elem("shortcuts")
const shortcutsB = elem("shortcuts-b")
const save = elem("save")
const publish = elem("publish")
var published = false

const ident = elem("name")
const school = elem("school")
const year = elem("year")
const month = elem("month")
const subject = elem("subject")
const theme = elem("theme")

const error = elem("error") 

const buttons = [rawB, previewB, shortcutsB]
const pages = [raw, preview, shortcuts]

const tab = "    "
const id = new URLSearchParams(window.location.search).get("id")

function input(e) {
    save.disabled = false
}

const elems = elemByClass("update")
for(var i in elems) {
    const e = elems[i]
    e.oninput = input
}

if(id != "") {
    request("privatenote", {id: id}).then(j => {
        const err = getErr(j)
        if(err) {
            error.innerHTML = err
        } else {
            var n = j.Note
            raw.value = n.Content
            ident.value = n.Name
            school.selectedIndex = n.School
            year.value = n.Year
            month.value = n.Month
            subject.value = n.Subject
            theme.value = n.Theme
            published = n.Published
            switchPublish()
        }
    })
}


var markdown = new Markdown(defaultColors)

request("config").then(j => {
    console.log(j)
    markdown = new Markdown(j.Cfg.Colors)
}).catch(e => {}).then(() => loadText("components/shortcuts.txt").then(t =>  shortcuts.innerHTML = markdown.convert(t)))

var undo = new Undo(4, 200) 

window.addEventListener("keydown", e => {
    if(!e.altKey) {
        return
    }

    switch(e.key) {
        case "r":
            rawB.click()
            break
        case "p":
            previewB.click()
            break
    }
    e.preventDefault()
})

buttons.forEach(b => {
    b.addEventListener("click", ev => {
        ev.preventDefault()
        pages.forEach(p => {
            p.hidden = true
        })
    })
})

rawB.onclick = function(ev) {
    ev.preventDefault()
    raw.hidden = false
    raw.select()
    raw.selectionStart = raw.selectionEnd
}

previewB.onclick = function(ev) {
    ev.preventDefault()
    preview.hidden = false
    preview.innerHTML = markdown.convert(raw.value)
    console.log(preview.innerHTML)
}

shortcutsB.onclick = function(ev) {
    ev.preventDefault()
    shortcuts.hidden = false
}

save.onclick = function(ev) {
    ev.preventDefault()
    if(!IsComplete()) {
        error.innerHTML = "you have to have all fields filled in to save, you can change them later but they have to be present"
        return
    }

    error.innerHTML = ""

    request("save", {
        name: ident.value, 
        school:school.value, 
        year:year.value, 
        subject:subject.value, 
        theme:theme.value, 
        month:month.value, 
        id:id
    }, {
        method: "POST",
        headers: {
            'content-type': 'text/plain'
        },
        body: raw.value,
    }).then(j => {
        const err = getErr(j)
        if(err) {
            error.innerHTML = err
        } else {
            save.disabled = true 
            console.log(j.ID)
            if(id !== j.ID) {
                window.location.href = `editor.html?id=${j.ID}`
            }
        }
    })
}

publish.onclick = function(ev) {
    ev.preventDefault()
    if(!published && !IsComplete) {
        error.innerHTML = "you cannot publish if some of fields are missing, they have to be present for optimal search"
    }

    if(id == "new") {
        error.innerHTML = "you have to save note first"
        return
    }

    error.innerHTML = ""

    request("setpublished", {id: id, publish: !published}).then( j => {
        const err = getErr(j)
        if(err) {
            error.innerHTML = err
        } else {
            published = !published
            switchPublish()
        }
    })
}

raw.addEventListener("keydown", e => {
    if(!e.altKey) {
        saveUndo()
        return
    }

    const start = raw.selectionStart, end = raw.selectionEnd

    if(markdown.set.has(e.key)) {
        const sub = `<${e.key}>`
        raw.value = insert(start, raw.value, sub)
        raw.value = insert(end + sub.length, raw.value, sub)
        raw.selectionStart = start + sub.length
        raw.selectionEnd = end + sub.length
        forceSaveUndo() 
        e.preventDefault()
        return
    }
    
    switch(e.key) {
        case "e":
            var s = raw.value.substring(start, end)
            markdown.set.forEach(element => {
                s = s.replaceAll(`<${element}>`, "")
            });
            raw.value = replace(start, end, raw.value, s)
            raw.selectionStart = start
            raw.selectionEnd = start + s.length
            forceSaveUndo()
            break
        case "z":
            if(undo.canUndo()){
                const s = undo.undo()
                console.log(s)
                raw.value = s.text
                raw.selectionStart = s.start
                raw.selectionEnd = s.end
            }
            break
        case "y":
            if(undo.canRedo()) {
                const s = undo.redo()
                raw.value = s.text
                raw.selectionStart = s.start
                raw.selectionEnd = s.end
            }
            break
        case "x":
        case "v":
            forceSaveUndo()
            break
        default:
            return
    }

    e.preventDefault()
})

var indent = 0

raw.addEventListener("keydown", e => {
    if(raw.selectionEnd != raw.selectionStart) return
    const pos = raw.selectionStart
    switch(e.key) {
        case "Tab":
            
            const tb = " ".repeat(tab.length - (pos - lineStart()) % tab.length)
            raw.value = insert(pos, raw.value, tb)
            raw.selectionStart = raw.selectionEnd = pos+tb.length
            e.preventDefault()
            return false
        case "ArrowRight":
            for(var i in  markdown.blocks) {
                const b = markdown.blocks[i]
                if(markdown.check(b, raw.selectionStart, raw.value)) {
                    raw.selectionStart += b.start.length
                    e.preventDefault() 
                    return false
                }
            }
            if(inFrontOfTab()) {
                select(pos+tab.length)
                e.preventDefault()
                return false
            }
            break
        case "ArrowLeft":
            if(behindTab()) {
                select(pos-tab.length)
                e.preventDefault()
                return false
            }
            break
        case "Enter":
            var res = 0
            for(var i = raw.selectionStart; i >= tab.length; i--) {
                const str = raw.value.substring(i-tab.length, i)
                if(str.includes("\n")) {
                    break
                }

                if(str == tab) {
                    res++
                    i-= tab.length - 1
                } else {
                    res = 0
                }

                
            }
            const str = "\n"+tab.repeat(res)
            raw.value = insert(pos, raw.value, str)
            raw.selectionStart = raw.selectionEnd = pos+str.length
            e.preventDefault()
            return false
        case "Backspace":
            if(behindTab()) {
                del(4)
                e.preventDefault()
                return false
            }
    }
})

function switchPublish() {
    publish.innerHTML = publish.innerHTML = (published) ? "take down" : "publish"
}

function IsComplete() {
    return IsFilled(ident, year, month, subject, theme)
}

function saveUndo() {
    undo.save(new UndoState(raw.value, raw.selectionStart, raw.selectionEnd))
}

function forceSaveUndo() {
    undo.forceSave(new UndoState(raw.value, raw.selectionStart, raw.selectionEnd))
}

function behindTab() {
    return raw.value.substring(raw.selectionStart - tab.length, raw.selectionStart) == tab && isTabAligned()
}

function inFrontOfTab() {
    return raw.value.substring(raw.selectionStart, raw.selectionStart + tab.length) == tab && isTabAligned()
}

function isTabAligned() {
    return (raw.selectionStart - lineStart()) % 4 == 0
}

function lineStart() {
    for(var i = raw.selectionStart; i >= 0; i--) {
        if(raw.value[i] == "\n") return i+1
    }

    return 0
}

function del(len) {
    const pos = raw.selectionStart
    raw.value = replace(pos - len, pos, raw.value, "")
    select(pos - len)
}

function write(str) {
    const pos = raw.selectionStart
    raw.value = insert(pos, raw.value, str)
    select(pos + str.length)
}

function select(idx) {
    raw.selectionStart = raw.selectionEnd = idx
}