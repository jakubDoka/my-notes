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

if(id != "new") {
    fetch(`/note?id=${id}`).then(e => {
        e.json().then(j => {
            if(j.Resp.Status != "success") {
                error.innerHTML = j.Resp.Status
            } else {
                var n = j.Note
                raw.value = n.Content.toString()
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
    })
}


var markdown = new Markdown(defaultColors)

fetch("/config").then(r => r.json()).then(j => {
    console.log(j)
    markdown = new Markdown((j.Cfg.Colors == null || j.Resp.Status != "success") ? defaultColors : j.Cfg.Colors)
}).catch(e => {}).then(() => {
    shortcuts.innerHTML = markdown.convert(
        `<2><t>Shortcuts<t><2>

<b>Alt<b>+<b>t<b> - creates <t>Title<t> text
<b>Alt<b>+<b>b<b> - creates <b>bold<b> text
<b>Alt<b>+<b>i<b> - creates <i>italic<i> text
<b>Alt<b>+<b>u<b> - creates <u>underlined<u> text
    
<b>Alt<b>+<b>(1-9)<b> - creates <3>colored<3> <2>text<2>, colors can be configured
    
<b>Alt<b>+<b>e<b> - erases stile tags within selection but leaves the text

<b>Alt<b>+<b>p<b> - switches to preview
<b>Alt<b>+<b>r<b> - switches to raw

<b>Alt<b>+<b>z<b> - undo
<b>Alt<b>+<b>y<b> - redo
<b>Alt<b>+<b>s<b> - save
    
its a <b><1>left<1><b> <b>Alt<b>

you can also <1><b><u><t>combine<t><u><b><1> stiles, though using title in sentence is equivalent of screaming`
    )
})

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

    fetch(`/save?name=${ident.value}&school=${school.value}&year=${year.value}&subject=${subject.value}&theme=${theme.value}&month=${month.value}&id=${id}`, {
        method: "POST",
        headers: {
            'content-type': 'text/plain'
        },
        body: raw.value,
    }).then(e => {
        e.json().then(e => {
            if(e.Resp.Status != "success") {
                error.innerHTML = e.Resp.Status
            } else {
                save.disabled = true 
                if(id != e.ID) {
                    window.location.href = `editor.html?id=${e.ID}`
                }
            }
        })
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

    fetch(`/setpublished?id=${id}&b=${!published}`).then(e => e.json().then( j => {
        if(j.Status != "success") {
            error.innerHTML = j.Status
        } else {
            published = !published
            switchPublish()
        }
    }))
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

raw.addEventListener("keydown", e => {
    switch(e.key) {
        case "Tab":
            const s = raw.selectionStart
            raw.value = raw.value.substring(0, s) + tab + raw.value.substring(s, raw.value.length)
            raw.selectionStart = raw.selectionEnd = s+tab.length
            e.preventDefault()
            return false
        case "ArrowRight":
            for(var i in  markdown.blocks) {
                const b = markdown.blocks[i]
                if(markdown.check(b, raw.selectionStart, raw.value)) {
                    raw.selectionStart += b.start.length
                    e.preventDefault() 
                    break
                }
            }
    }
})

function switchPublish() {
    publish.innerHTML = publish.innerHTML = (published) ? "take down" : "publish"
}

function IsComplete() {
    return IsFillet(ident, year, month, subject, theme)
}

function saveUndo() {
    undo.save(new UndoState(raw.value, raw.selectionStart, raw.selectionEnd))
}

function forceSaveUndo() {
    undo.forceSave(new UndoState(raw.value, raw.selectionStart, raw.selectionEnd))
}