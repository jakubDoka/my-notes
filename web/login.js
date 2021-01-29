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

    if(confirm != password) {
        err.innerHTML = "confirm password does not match"
        return
    }
    
    err.innerHTML = ""
    sha256(password).then((str)=>{
        fetch(`/register?n=${nm.value}&p=${str}&e=${email.value}`).then(re=>re.json()).then(j => {
            if (j.Status != "success") {
                err.innerHTML = j.Status
            } else {
                err.innerHTML = `account successfully created, we sent you verification email,
                 enter the code and press verify, then you can login`
            }
        })
    })
}

verify.onclick = function(ev) {
    ev.preventDefault()
    if (missingName() || missingPassword() || missingCode()) {
        return
    }

    sha256(password).then((str)=>{
        fetch(`/verify?n=${nm.value}&p=${str}&c=${code.value}`).then((re)=>{
            re.json().then((j) => {
                if (j.Status != "success") {
                    err.innerHTML = j.Status
                } else {
                    err.innerHTML = "your email was verified, you can now login"
                }
            })
        })
    })

}

login.onclick = function(ev) {
    ev.preventDefault()
    if (missingName() || missingPassword()) {
        return
    }

    sha256(password).then((str)=>{
        fetch(`/login?n=${nm.value}&p=${str}`).then((re)=>{
            re.json().then((j) => {
                if (j.Status != "success") {
                    elem("error").innerHTML = j.Status
                } else {
                    alert(`we used one cookie to store your user information so you don't 
                    have to login anytime you do anything, the cookie will be deleted when 
                    you close a browser so you then have to er-log, to delete cookie use logout`)
                    window.location.href = "account.html"
                }
            })
        })
    })
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