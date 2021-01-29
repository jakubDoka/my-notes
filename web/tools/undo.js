class Undo {
    constructor(saveFrequency, cap) {
        this.undoStack = []
        this.redoStack = []
        this.saveFrequency = saveFrequency
        this.cap = cap
        this.counter = 0
    }

    canUndo() {
        return this.undoStack.length != 0
    }

    canRedo() {
        return this.redoStack.length != 0
    }

    undo() {
        const stp = this.undoStack.pop()
        this.redoStack.push(stp)
        this.truncate()
        return stp
    }

    redo() {
        const stp = this.redoStack.pop()
        this.undoStack.push(stp)
        this.truncate()
        return stp
    }

    save(state) {
        this.counter += 1
        if (this.saveFrequency > this.counter) {
            return
        }
        this.counter = 0
        this.forceSave(state)
    }

    truncate() {
        if(this.undoStack.length > this.cap) {
            this.undoStack.splice(0, 1)
        } else if(this.redoStack.length > this.cap) {
            this.redoStack.splice(0, 1)
        }
    }

    forceSave(state) {
        if(this.canUndo() && state == this.undoStack[this.undoStack.length-1]) {
            return
        }
        this.undoStack.push(state)
        this.truncate()
    }
}