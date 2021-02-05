const nm = elem("name")
const password = elem("password")
const confirm = elem("confirm")
const email = elem("email")
const code = elem("code")

const err = elem("error")

const login = elem("login")
const singIn = elem("sing-in")
const verify = elem("verify")
const logout = elem("logout")

logout.onclick = function(ev) {
    ev.preventDefault()
    document.cookie = ""
    err.innerHTML = "all cookies were deleted"
}

singIn.onclick = function(ev) {
    ev.preventDefault()
    if (missingName() || missingPassword() || missingEmail()) {
        return
    }

    if(!IsValidName(nm.value)) {
        err.innerHTML = invalidNameMessage
        return
    }

    if(confirm.value != password.value) {
        err.innerHTML = "confirm password does not match"
        return
    }
    
    err.innerHTML = ""
    sha256(password).then((str)=> request("register", {name: nm.value, password: str, email: email.value}).then(j => {
        const err2 = getErr(j)
        if (err2) {
            err.innerHTML = err2
        } else {
            err.innerHTML = `account successfully created, we sent you verification email,
                enter the code and press verify, then you can login`
        }
    }))
}

verify.onclick = function(ev) {
    ev.preventDefault()
    if (missingName() || missingPassword() || missingCode()) {
        return
    }

    sha256(password).then((str)=> request("verify", {name: nm.value, password: str, code: code.value}).then((j) => {
        const err2 = getErr(j)
        if (err2) {
            err.innerHTML = err2
        } else {
            err.innerHTML = "your email was verified, you can now login"
        }
    }))

}

login.onclick = function(ev) {
    ev.preventDefault()
    if (missingName() || missingPassword()) {
        return
    }

    sha256(password).then((str)=> request("login", {name: nm.value, password: str}).then((j) => {
        const err2 = getErr(j)
        if (err2) {
            err.innerHTML = err2
        } else {
            window.location.href = "account.html"
        }
    }))
}

function missingName() {
    if(nm.value == "") {
        err.innerHTML = "name is necessary"
    } 

    return nm.value == ""
}

function missingPassword() {
    if(password == "") {
        err.innerHTML = "missing password"
        return true
    } 
    
    return false
}

function missingEmail() {
    if(email.value == "") {
        err.innerHTML = "you haven't provided email"
    } 
    
    return email.value == ""
}

function missingCode() {
    if(code.value == "") {
        err.innerHTML = "you did not enter any verification code"
    }

    return code.value == ""
}