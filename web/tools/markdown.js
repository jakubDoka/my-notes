class Block {
    constructor(start, cl) {
        this.raw = start
        this.start = `<${start}>`
        this.cl = cl
        this.color = ""
    }

    Start() {
        if (this.color == "") {
            return `<span class=${this.cl}>`
        }
        return `<span class="base ${this.cl}" style="color: ${this.color};">`
    }
}

function Colored(start, color) {
    b = new Block(start, "")
    b.color = color
    return b 
}

const defaultColors = ["#b03830", "#b0972a", "#5d62f0"] 

class Markdown {
    constructor(colors) {
        if(colors == null) {
            colors = defaultColors
        }

        this.blocks = [
            new Block("t", "title"),
            new Block("b", "bold"),
            new Block("i", "italic"),
            new Block("u", "underline"),
        ]

        colors.forEach((c, i) => {
            this.blocks.push(Colored((i+1).toString(), c))
        })

        this.set = new Set()

        this.blocks.forEach( b => {
            this.set.add(b.raw)
        })

    }

    check(b, idx, str) {
        return b.start == str.substring(idx, idx+b.start.length)
    }

    convert(raw) {
        var result = []
        var stack = []
        var last = 0
        var i = 0

        var handle = function(b, end) {
            result.push(raw.substring(last, i), (end) ? "</span>" : b.Start())
            i += b.start.length-1
            last = i+1
        }

        for(; i < raw.length; i++) {
            if(stack.length != 0 && this.check(stack[stack.length-1], i, raw)) {
                handle(stack.pop(), true)
                continue
            }
            for(var j in this.blocks) {
                const b = this.blocks[j]
                if(this.check(b, i, raw)) {
                    handle(b, false)
                    stack.push(b)
                }
            }
        }
    
        result.push(raw.substring(last))
    
        for(b in stack){
            result.push("</span>")
        }
    
        return result.join("").replaceAll("\n", "<br><hr>").replaceAll("    ", "<tab></tab>")
    }
}