
const allCookies = document.cookie;
const dataContainer = document.getElementById("data-container");

// Get the Schema for the queries
fetch('http://localhost:8080/schema', {
    method: 'GET', // or 'GET', 'PUT', 'DELETE'
    headers: {
        'Content-Type': 'application/json'
    },
    // body: JSON.stringify({ key: 'value' }) // Only for methods with body (e.g., POST, PUT)
})
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then(schemaData => {
        const dataTypes = document.getElementById("dataTypes");
        dataTypes.replaceChildren();
        const getQueries = document.getElementById("getQueries");
        getQueries.replaceChildren();
        for (var dataType in schemaData) {
            const dataTypeOpt = document.createElement("option");
            dataTypeOpt.value = dataType;
            dataTypeOpt.innerText = dataType;
            dataTypes.appendChild(dataTypeOpt);

            const queries = schemaData[dataType].queries;
            for (var qIdx = 0; qIdx < queries.length; qIdx++) {
                const query = queries[qIdx];
                const queryContainer = document.createElement("div");
                queryContainer.className = dataType;
                queryContainer.id = query;
                queryContainer.style = "display: none;";
                const queryLabel = document.createElement("label");
                queryLabel.for = query;
                queryLabel.innerText = query;
                const queryInput = document.createElement("input");
                queryInput.type = "text";
                queryInput.name = query;
                queryContainer.appendChild(queryLabel);
                queryContainer.appendChild(queryInput);
                getQueries.appendChild(queryContainer);
            }
        }
        selectDataType();
    })
    .catch(error => {
        console.error('Error:', error);
    });

// Get the user data
fetch('http://localhost:8080/data/user', {
    method: 'GET', // or 'GET', 'PUT', 'DELETE'
    headers: {
        'Content-Type': 'application/json'
    },
    // body: JSON.stringify({ key: 'value' }) // Only for methods with body (e.g., POST, PUT)
})
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        document.getElementById("data-container").style = "display: block;";
        pfp = document.getElementById("pfp");
        pfp.replaceChildren();
        pinfo = document.getElementById("pinfo");
        pinfo.replaceChildren();
        if (data.length > 0) {
            const pfpElement = document.createElement("img");
            pfpElement.src = data[0].picture;
            pfpElement.alt = "profile picture";
            pfp.appendChild(pfpElement);
            const pinfoName = document.createElement("p");
            pinfoName.innerText = "Name: " + data[0].name;
            const pinfoEmail = document.createElement("p");
            pinfoEmail.innerText = "Email: " + data[0].email;
            pinfo.appendChild(pinfoName);
            pinfo.appendChild(pinfoEmail);
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });

// Send requests on button click
const button = document.getElementById('myButton');
button.addEventListener('click', () => {
    const requestMethod = getSelectedMethod();
    const requestUrl = buildUrl(requestMethod);
    const request = buildRequest(requestMethod);
    const responseArea = document.getElementById("responseData");
    responseArea.replaceChildren();
    fetch(requestUrl, request)
        .then(response => {
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            return response.json();
        })
        .then(data => {
            // console.log('Success:', data);
            const responseArea = document.getElementById("responseData");
            const dataString = JSON.stringify(data, null, 4);
            responseArea.innerText = dataString;
        })
        .catch(error => {
            console.error('Error:', error);
        });
});

// Request builders
function buildRequest(requestMethod) {
    var request = {
        method: requestMethod,
        headers: {
            'Content-Type': 'application/json'
        }
    };
    if (requestMethod === "POST" || requestMethod === "PUT") {
        request.body = getPrimaryInput();
    }
    return request;
}

function buildUrl(method) {
    const dataType = selectDataType();
    const queryParams = getActiveQueryValues();
    const idString = getPrimaryInput();
    var url = "/data/" + dataType;
    if (idString.length > 0 && (method === "GET" || method === "DELETE")) {
        url += "/" + idString;
    }
    if (queryParams.length > 0) {
        url += queryParams;
    }
    return url;
}

// Input helpers
function getPrimaryInput() {
    const methods = document.getElementById("requestMethodSelection");
    const active = methods.getElementsByClassName("tabcontent active");
    return "" + active[0].getElementsByTagName('input')[0].value;
}

function getSelectedMethod() {
    const methods = document.getElementById("requestMethodSelection");
    const active = methods.getElementsByClassName("tabcontent active");
    if (active.length > 0) {
        return active[0].id;
    }
    return "GET";
}

function getActiveQueryValues() {
    const selectObj = document.getElementById("dataTypes");
    var idx = selectObj.selectedIndex;
    var dataType = selectObj[idx].value;
    const getQueries = document.getElementById("getQueries");
    const dataTypeQueries = getQueries.getElementsByClassName(dataType);
    var queriesString = "";
    var activeQueryIdx = 0;
    for (var qIdx = 0; qIdx < dataTypeQueries.length; qIdx++) {
        const input = dataTypeQueries[qIdx].getElementsByTagName('input')[0];
        const queryName = input.name;
        const queryVal = input.value;
        if (queryVal.length > 0) {
            if (activeQueryIdx === 0) {
                queriesString += "?";
            } else {
                queriesString += "&";
            }
            queriesString += input.name + "=" + input.value;
            activeQueryIdx += 1;
        }
    }
    return queriesString;
}

// DOM manipulation functions
function selectDataType() {
    const selectObj = document.getElementById("dataTypes");
    var idx = selectObj.selectedIndex;
    var dataType = selectObj[idx].value;
    const getQueries = document.getElementById("getQueries");
    const getQueriesChildren = getQueries.children;
    for (var qIdx = 0; qIdx < getQueriesChildren.length; qIdx++) {
        getQueriesChildren[qIdx].style.display = "none";
    }
    const dataTypeQueries = getQueries.getElementsByClassName(dataType);
    for (var qIdx = 0; qIdx < dataTypeQueries.length; qIdx++) {
        dataTypeQueries[qIdx].style.display = "block";
    }
    return dataType;
}

function openMethod(evt, methodName) {
    var i, tabcontent, tablinks;
    tabcontent = document.getElementsByClassName("tabcontent");
    for (i = 0; i < tabcontent.length; i++) {
        tabcontent[i].style.display = "none";
        tabcontent[i].className = tabcontent[i].className.replace(" active", "");
    }
    tablinks = document.getElementsByClassName("tablinks");
    for (i = 0; i < tablinks.length; i++) {
        tablinks[i].className = tablinks[i].className.replace(" active", "");
    }
    document.getElementById(methodName).style.display = "block";
    document.getElementById(methodName).className += " active";
    evt.currentTarget.className += " active";
}
