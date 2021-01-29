{
    const target = document.currentScript.getAttribute("target")
    const path = document.currentScript.getAttribute("source")
    const e = document.getElementById(target)
    
    fetch(path).then(file => file.text()).then(text => {
        e.innerHTML = text
        Rerun(e)
    })
    
    function Rerun(elem) {
        elem.childNodes.forEach(element => {
            if(element.tagName == "SCRIPT") {
                var data = (element.text || element.textContent || element.innerHTML || "" ),
                head = document.getElementsByTagName("head")[0] ||
                          document.documentElement,
                script = document.createElement("script");
        
                script.type = "text/javascript";
                try {
                // doesn't work on ie...
                script.appendChild(document.createTextNode(data));      
                } catch(e) {
                // IE has funky script nodes
                script.text = data;
                }
            
                head.insertBefore(script, head.firstChild);
                head.removeChild(script);
            }
            Rerun(element)
        })
    }
}

