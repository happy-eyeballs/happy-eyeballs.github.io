function getRand() {
    const randomArray = new Uint32Array(1);
    self.crypto.getRandomValues(randomArray);
    return randomArray[0];
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

function getRandomRunId() {
    const max = 100000;
    const min = 0;
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

async function checkv6(delay, randid) {
    const url = `https://id-${randid}.delay-${delay}.v1.example.com:10443/ping`;
    console.log(`measuring id ${randid}`);
    let isv6 = await fetch(url, { cache: 'no-store' })
        .then(response => {
            if (!response.ok) {
                console.log(`error for ${url}`);
            } else {
                return response.text();
            }
        }).then(responsetext => {
            console.log(`response is: ${responsetext}`);
            return responsetext.includes(':')
        }).catch((error) => {
            return error
        });
    if (typeof(isv6) !== 'boolean') {
        return {"error": isv6, "uniqueId": randid}
    }
    return {"isV6": isv6, "uniqueId": randid, "error": null};
}


async function checkv6ForV2(delay, randid, delayType) {
    const url = `https://v2delay_${delayType}-${randid}_${delay}.v2.example.com:12443/ping`;
    console.log(`measuring id ${randid} ${delayType} ${delay}`);
    let isv6 = await fetch(url, { cache: 'no-store' })
        .then(response => {
            if (!response.ok) {
                console.log(`error for ${url}`);
            } else {
                return response.text();
            }
        }).then(responsetext => {
            console.log(`response is: ${responsetext}`);
            return responsetext.includes(':')
        }).catch((error) => {
            return error
        });
    if (typeof(isv6) !== 'boolean') {
        return {"error": isv6, "uniqueId": randid, "delayType": delayType}
    }
    return {"isV6": isv6, "uniqueId": randid, "delayType": delayType, "error": null};
}

let configuredDelays = [];

export async function setup() {
    await fetch("delays.csv", { cache: 'no-store' })
        .then(response => {
            if (!response.ok) {
                console.log(`error for ${url}`);
            } else {
                return response.text();
            }
        }).then(responsetext => {
            console.log(`response is: ${responsetext}`);
            responsetext.split('\n').forEach(line => {
                if (line.trim() !== '') {
                    configuredDelays.push(line.trim());
                }
            });
        });

    if (resultsPath != "v2results") {
        configuredDelays.pop();
        configuredDelays.pop();
    }
    let tableHeader = document.getElementById("testRunTableHeader");
    let delayHeader = document.createElement("th");
    delayHeader.textContent = "Delays [ms]";
    delayHeader.setAttribute("colspan", configuredDelays.length);
    tableHeader.appendChild(delayHeader);

    let tableSubHeader = document.getElementById("testRunTableHeaderDelays");
    let delayHeaders = [];

    configuredDelays.forEach(delay => {
        let element = document.createElement("th");
        element.style.verticalAlign = "bottom";
        element.style.alignContent = "start";

        // element.setAttribute("class", "position-relative")
        let elementdiv = document.createElement("div");
        element.appendChild(elementdiv);

        // elementdiv.style.transform = "rotate(270deg)"
        // elementdiv.setAttribute("class", "position-absolute top-0 start-40")
        elementdiv.textContent = `${delay}`;
        delayHeaders.push(element);
    });
    delayHeaders.forEach(element => {
        tableSubHeader.appendChild(element);
    });

    console.log(configuredDelays);
}

let userFormIds = []
let testRunCount = 0;
const testRunData = {};
let testIDsTransmitted = [];
let resultsPath = "";

export function setUserFormIds(newUserFormIds) {
    userFormIds = newUserFormIds;
}

export function setResultsPath(newResultsPath) {
    resultsPath = newResultsPath;
}

export async function measureHappyEyeballs() {
    const transmitEnabled = document.getElementById('transmitResultsBtn').getAttribute('disabled') == null
    disableUserInteraction()

    const repetitions = Number(document.getElementById('repetitions').value);
    const randomizeDomains = document.getElementById('domainRandomization').checked;

    // check if ipv4 and ipv6 is available
    const v4address = await getAddress(false);
    const v6address = await getAddress(true);
    if (v4address == null) {
        alert('No IPv4 Address available');
        enableUserInteraction();
        if (!transmitEnabled) {
            document.getElementById('transmitResultsBtn').setAttribute('disabled', true);
        }
        return;
    }
    if (v6address == null) {
        alert('No IPv6 Address available');
        enableUserInteraction();
        if (!transmitEnabled) {
            document.getElementById('transmitResultsBtn').setAttribute('disabled', true);
        }
        return;
    }

    // Increment test run count
    testRunCount++;
    const runId = getRand();
    const startDate = new Date();
    const runInfo = {
        "id": runId,
        "runCount": testRunCount,
        "timestampStart": startDate.getTime(),
        "userAgent": window.navigator.userAgent,
        "platform": window.navigator.platform,
        "vendor": window.navigator.vendor,
        "domainRandomization": randomizeDomains,
        "repetitions": repetitions,
        "client_addr_v4": v4address.replace(/(\r\n|\n|\r)/gm, ""),
        "client_addr_v6": v6address.replace(/(\r\n|\n|\r)/gm, ""),
        "userInfo": document.getElementById('userInfo').value,
        "delayResults": []};

    const runUIdMapping = {};
    if (!randomizeDomains) {
        for (const delay of configuredDelays) {
            runUIdMapping[delay] = getRandomRunId();
        }
    }

    let infoElement = document.getElementById("testInfo");
    let infoElementText = document.getElementById("testInfoText");

    infoElement.classList.remove("d-none");

    for (var i = 0; i < repetitions; i++) {
        const delayElements = createResultRow(testRunCount, i, repetitions, randomizeDomains, '');

        await executeTestRun(runInfo, delayElements, runUIdMapping, randomizeDomains, infoElementText, i, repetitions);

        if (i + 1 < repetitions) {
            infoElementText.textContent = `Sleeping for 5s between Test Runs (Run ${i + 1} of ${repetitions})`
            await sleep(5000);
        }
    }

    testRunData[runId] = runInfo;

    infoElement.classList.add("d-none");

    enableUserInteraction()

    if (document.getElementById('autoTransmit').checked) {
        transmitResults()
    }
}

function disableUserInteraction() {
    for (const formId of userFormIds) {
        document.getElementById(formId).setAttribute('disabled', true);
    }
}

function enableUserInteraction() {
    for (const formId of userFormIds) {
        document.getElementById(formId).removeAttribute('disabled');
    }
}

async function getAddress(ipv6) {
    let version = 4;
    if (ipv6) {
        version = 6;
    }
    const url = `https://ipv${version}-only.v1.example.com:11443/my-ip`;
    let address = await fetch(url, { cache: 'no-store' })
        .then(response => {
            if (!response.ok) {
                console.log(`error for ${url}`);
                return null;
            } else {
                return response.text();
            }
        }).catch((error) => {
            return null;
        });
    return address;
}

function createResultRow(testRunCount, repetition, repetitions, randomizeDomains, addendum) {
    // Create new row
    const newRow = document.createElement("tr");

    // Create cells
    const cellRunNumber = document.createElement("td");
    const cellTimestamp = document.createElement("td");

    // Populate cells
    cellRunNumber.textContent = `${testRunCount} (${repetition + 1}/${repetitions})`;
    if (addendum) {
        cellRunNumber.textContent = `${cellRunNumber.textContent} ${addendum}`
    }
    cellTimestamp.textContent = new Date().toLocaleString();

    if (randomizeDomains) {
        var templateImg = document.querySelector('#randomizeImg');
    } else {
        var templateImg = document.querySelector('#noRandomizeImg');
    }
    const clone = templateImg.content.cloneNode(true);
    cellRunNumber.appendChild(clone);

    newRow.appendChild(cellRunNumber);
    newRow.appendChild(cellTimestamp);

    let delayElements = {};
    for (const delay of configuredDelays) {
        let element = document.createElement("td");

        delayElements[delay] = element;
        newRow.appendChild(element);
    }

    document.getElementById("testRunTableBody").appendChild(newRow);
    return delayElements;
}

async function executeTestRun(runInfo, delayElements, runUIdMapping, randomizeDomains, infoElementText, repetition, totalRepetitions) {
    const timingInfos = {};
    const v6Infos = {};
    try {
        const observer = new PerformanceObserver((list) => {
            const entries = list.getEntriesByType('resource');
            entries.forEach((entry) => {
                if (entry.name.endsWith("v1.example.com:10443/ping")) {
                    const responseTime = entry.duration;
                    const urlIdEndIndex = entry.name.search(/\./);
                    if (urlIdEndIndex === -1) {
                        console.log(`Could not find id in ${entry.name}`);
                    } else {
                        const id = entry.name.substring(11, urlIdEndIndex); // subtract https://id-
                        console.log(`--- Timings for ${id} - ${entry.name} ${responseTime.toFixed(2)} ms ---`);
                        timingInfos[id] = responseTime.toFixed(2);
                    }
                }
            });
        });

        observer.observe({ type: 'resource', buffered: true });

        for (const delay of configuredDelays) {
            infoElementText.textContent = `Checking for delay of ${delay}ms (Run ${repetition + 1} of ${totalRepetitions})`;
            if (randomizeDomains) {
                var runUId = getRandomRunId();
            } else {
                var runUId = runUIdMapping[delay];
            }

            const runTimestamp = new Date().getTime();
            const isV6Result = await checkv6(delay, runUId);
            if (isV6Result["error"] != null) {
                v6Infos[delay] = {"delay": delay, "runUId": isV6Result.uniqueId, "error": true, "timestamp": runTimestamp, "repetition": repetition};
            } else {
                v6Infos[delay] = {"delay": delay, "runUId": isV6Result.uniqueId, "isV6": isV6Result.isV6, "timestamp": runTimestamp, "repetition": repetition, "error": false};
            }
            let element = delayElements[delay];
            if (isV6Result["error"] != null) {
                element.setAttribute("class", "bg-danger");
                let subelement = document.createElement("span");
                subelement.textContent = 'err';
                element.appendChild(subelement);
            } else if (isV6Result.isV6) {
                element.setAttribute("class", "bg-success");
                let subelement = document.createElement("span");
                subelement.textContent = 'v6';
                element.appendChild(subelement);
            } else {
                element.setAttribute("class", "bg-warning");
                let subelement = document.createElement("span");
                subelement.textContent = 'v4';
                element.appendChild(subelement);
            }
        }

        await sleep(500);
        observer.disconnect();
    } catch (error) {
        console.error('Error measuring network timings:', error);
    }

    for (const delay of configuredDelays) {
        const result = v6Infos[delay];
        result["responseTime"] = timingInfos[String(result['runUId'])];
        runInfo["delayResults"].push(result);
    }

    const endDate = new Date();
    runInfo["timestampEnd"] = endDate.getTime();
}

export async function transmitResults() {
    document.getElementById("startTestBtn").setAttribute('disabled', true);
    document.getElementById("transmitResultsBtn").setAttribute('disabled', true);

    const dataToPush = [];
    for (const runID of Object.keys(testRunData)) {
        if (testIDsTransmitted.includes(runID)) {
            continue;
        }
        dataToPush.push(testRunData[runID]);
    }

    fetch(resultsPath, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(dataToPush)
    }).then((response) => {
        if (!response.ok) {
            alert(`Could not transmit the results ${response.status}. Check JS Console for more Info`);
            throw new Error(`HTTP error! Status: ${response.status} for ${JSON.stringify(dataToPush)}`);
        }
    });

    for (const runInfo of dataToPush) {
        testIDsTransmitted.push(String(runInfo["id"]));
    }

    var templateImg = document.querySelector('#uploadedImg');
    const tbody = document.getElementById('testRunTableBody')
    for (const row of tbody.children) {
        let hasUploadImage = false;
        for (const child of row.children[0].children) {
            if (child.id === "uploadedImgdiv") {
                hasUploadImage = true;
            }
        }
        if (!hasUploadImage) {
            const clone = templateImg.content.cloneNode(true);
            row.children[0].appendChild(clone)
        }
    }

    document.getElementById("startTestBtn").removeAttribute('disabled');
    // document.getElementById("transmitResultsBtn").removeAttribute('disabled');
}

export async function measureHappyEyeballsDNS() {
    const transmitEnabled = document.getElementById('transmitResultsBtn').getAttribute('disabled') == null
    disableUserInteraction()

    const repetitions = Number(document.getElementById('repetitions').value);
    const randomization = document.getElementById('runRandomization').checked;

    if (document.getElementById('resolverInfo').value == "") {
        alert("Resolver information needed!");
        enableUserInteraction();
        if (!transmitEnabled) {
            document.getElementById('transmitResultsBtn').setAttribute('disabled', true);
        }
        return;
    }

    // Increment test run count
    testRunCount++;
    const runId = getRand();
    const startDate = new Date();
    const runInfo = {
        "id": runId,
        "runCount": testRunCount,
        "timestampStart": startDate.getTime(),
        "repetitions": repetitions,
        "resolverInfo": document.getElementById('resolverInfo').value,
        "resolutionInfos": []
    };

    let infoElement = document.getElementById("testInfo");
    let infoElementText = document.getElementById("testInfoText");

    infoElement.classList.remove("d-none");

    let runUId = getRandomRunId();
    for (var i = 0; i < repetitions; i++) {
        const delayElements = createResultRow(testRunCount, i, repetitions, false, '');

        await executeDNSTestRun(runInfo, delayElements, runUId, infoElementText, i, repetitions);

        if (i + 1 < repetitions) {
            infoElementText.textContent = `Sleeping for 5s between Test Runs (Run ${i + 1} of ${repetitions})`
            if (randomization) {
                var tempid = getRandomRunId();
                runUId = tempid;
            }
            await sleep(5000);
        }
    }

    testRunData[runId] = runInfo;

    infoElement.classList.add("d-none");

    enableUserInteraction()

    if (document.getElementById('autoTransmit').checked) {
        transmitResults()
    }
}

function createDNSResultRow(testRunCount, repetition, repetitions) {
    // Create new row
    const newRow = document.createElement("tr");

    // Create cells
    const cellRunNumber = document.createElement("td");
    const cellTimestamp = document.createElement("td");

    // Populate cells
    cellRunNumber.textContent = `${testRunCount} (${repetition + 1}/${repetitions})`;
    cellTimestamp.textContent = new Date().toLocaleString();

    newRow.appendChild(cellRunNumber);
    newRow.appendChild(cellTimestamp);

    delayElements = {};
    for (const delay of configuredDelays) {
        element = document.createElement("td");

        delayElements[delay] = element;
        newRow.appendChild(element);
    }

    document.getElementById("testRunTableBody").appendChild(newRow);
    return delayElements;
}


async function executeDNSTestRun(runInfo, delayElements, runUId, infoElementText, repetition, totalRepetitions) {
    const timingInfos = {};
    const v6Infos = {};
    try {
        const observer = new PerformanceObserver((list) => {
            const entries = list.getEntriesByType('resource');
            entries.forEach((entry) => {
                if (entry.name.endsWith("v1-rdns.example.com:10443/ping")) {
                    const responseTime = entry.duration;
                    const urlIdEndIndex = entry.name.search(/\./);
                    if (urlIdEndIndex === -1) {
                        console.log(`Could not find id in ${entry.name}`);
                    } else {
                        const id = entry.name.substring(11, urlIdEndIndex); // subtract https://id-
                        console.log(`--- Timings for ${id} - ${entry.name} ${responseTime.toFixed(2)} ms ---`);
                        timingInfos[id] = responseTime.toFixed(2);
                    }
                }
            });
        });

        observer.observe({ type: 'resource', buffered: true });

        for (const delay of configuredDelays) {
            infoElementText.textContent = `Checking for delay of ${delay}ms (Run ${repetition + 1} of ${totalRepetitions})`;

            const runTimestamp = new Date().getTime();
            let dnsResult = await performDNSRequest(delay, runUId);
            v6Infos[delay] = {"delay": delay, "runUId": dnsResult.uniqueId, "timestamp": runTimestamp, "repetition": repetition, "result": dnsResult.response};

            let element = delayElements[delay];
            element.setAttribute("class", "bg-info");
            let subelement = document.createElement("span");
            // subelement.textContent = '';
            element.appendChild(subelement);
        }

        await sleep(500);
        observer.disconnect();
    } catch (error) {
        console.error('Error measuring network timings:', error);
    }

    for (const delay of configuredDelays) {
        const result = v6Infos[delay];
        result["responseTime"] = timingInfos[String(result['runUId'])];
        runInfo["resolutionInfos"].push(result);
    }

    const endDate = new Date();
    runInfo["timestampEnd"] = endDate.getTime();
}

async function performDNSRequest(delay, randid) {
    const url = `https://id-${randid}.dns-delay-${delay}.v1-rdns.example.com:10443/ping`;
    console.log(`measuring id ${randid}`);
    let response = await fetch(url, { cache: 'no-store' })
        .then(response => {
            if (!response.ok) {
                console.log(`error for ${url}`);
            } else {
                return response.text();
            }
        }).then(responsetext => {
            console.log(`response is: ${responsetext}`);
            return responsetext
        }).catch((error) => {
            return error
        });
    return {"response": response, "uniqueId": randid};
}


export async function measureHappyEyeballsV2() {
    const transmitEnabled = document.getElementById('transmitResultsBtn').getAttribute('disabled') == null
    disableUserInteraction()

    const repetitions = Number(document.getElementById('repetitions').value);
    const randomizeDomains = document.getElementById('domainRandomization').checked;

    // check if ipv4 and ipv6 is available
    const v4address = await getAddress(false);
    const v6address = await getAddress(true);
    if (v4address == null) {
        alert('No IPv4 Address available');
        enableUserInteraction();
        if (!transmitEnabled) {
            document.getElementById('transmitResultsBtn').setAttribute('disabled', true);
        }
        return;
    }
    if (v6address == null) {
        alert('No IPv6 Address available');
        enableUserInteraction();
        if (!transmitEnabled) {
            document.getElementById('transmitResultsBtn').setAttribute('disabled', true);
        }
        return;
    }

    // Increment test run count
    testRunCount++;
    const runId = getRand();
    const startDate = new Date();
    const runInfo = {
        "id": runId,
        "runCount": testRunCount,
        "timestampStart": startDate.getTime(),
        "userAgent": window.navigator.userAgent,
        "platform": window.navigator.platform,
        "vendor": window.navigator.vendor,
        "domainRandomization": randomizeDomains,
        "repetitions": repetitions,
        "client_addr_v4": v4address.replace(/(\r\n|\n|\r)/gm, ""),
        "client_addr_v6": v6address.replace(/(\r\n|\n|\r)/gm, ""),
        "userInfo": document.getElementById('userInfo').value,
        "delayResults": []};

    const runUIdMapping = {};
    if (!randomizeDomains) {
        for (const delay of configuredDelays) {
            runUIdMapping[delay] = getRandomRunId();
        }
    }

    let infoElement = document.getElementById("testInfo");
    let infoElementText = document.getElementById("testInfoText");

    infoElement.classList.remove("d-none");

    for (var i = 0; i < repetitions; i++) {
        const delayElements = {
            'a': createResultRow(testRunCount, i, repetitions, randomizeDomains, 'Delay A'),
            'aaaa': createResultRow(testRunCount, i, repetitions, randomizeDomains, 'Delay AAAA')
        };

        await executeV2TestRun(runInfo, delayElements, runUIdMapping, randomizeDomains, infoElementText, i, repetitions);

        if (i + 1 < repetitions) {
            infoElementText.textContent = `Sleeping for 5s between Test Runs (Run ${i + 1} of ${repetitions})`
            await sleep(5000);
        }
    }

    testRunData[runId] = runInfo;

    infoElement.classList.add("d-none");

    enableUserInteraction()

    if (document.getElementById('autoTransmit').checked) {
        transmitResults()
    }
}

async function executeV2TestRun(runInfo, delayElements, runUIdMapping, randomizeDomains, infoElementText, repetition, totalRepetitions) {
    const timingInfos = {'a': {}, 'aaaa': {}};
    const v6Infos = {'a': {}, 'aaaa': {}};
    try {
        const observer = new PerformanceObserver((list) => {
            const entries = list.getEntriesByType('resource');
            entries.forEach((entry) => {
                if (entry.name.endsWith("v2.example.com:10443/ping")) {
                    const responseTime = entry.duration;
                    const urlIdEndIndex = entry.name.search(/\./);
                    if (urlIdEndIndex === -1) {
                        console.log(`Could not find id in ${entry.name}`);
                    } else {
                        const isAAAA = entry.name.search(/AAAA/);
                        let delayType;
                        let substr;
                        if (isAAAA !== -1) {
                            delayType = 'a'
                            substr = entry.name.substring(18, urlIdEndIndex); // subtract https://v2delay_a-
                        } else {
                            delayType = 'aaaa'
                            substr = entry.name.substring(21, urlIdEndIndex); // subtract https://v2delay_aaaa-
                        }
                        const idEndIndex = entry.name.search(/-/);
                        if (idEndIndex === -1) {
                            console.log(`Could not find id in ${entry.name}`);
                        }

                        const id = substr.substring(0, idEndIndex);

                        console.log(`--- Timings for ${id} - ${entry.name} ${responseTime.toFixed(2)} ms ---`);
                        timingInfos[delayType][id] = responseTime.toFixed(2);
                    }
                }
            });
        });

        observer.observe({ type: 'resource', buffered: true });

        for (const delayType of ['a', 'aaaa']) {
            for (const delay of configuredDelays) {
                infoElementText.textContent = `Checking for delay of ${delay}ms (Run ${repetition + 1} of ${totalRepetitions})`;
                if (randomizeDomains) {
                    var runUId = getRandomRunId();
                } else {
                    var runUId = runUIdMapping[delay];
                }

                const runTimestamp = new Date().getTime();
                const isV6Result = await checkv6ForV2(delay, runUId, delayType);
                if (isV6Result["error"] != null) {
                    v6Infos[delayType][delay] = {"delay": delay, "runUId": isV6Result.uniqueId, "delayType": delayType, "error": true, "timestamp": runTimestamp, "repetition": repetition};
                } else {
                    v6Infos[delayType][delay] = {"delay": delay, "runUId": isV6Result.uniqueId, "delayType": delayType, "isV6": isV6Result.isV6, "timestamp": runTimestamp, "repetition": repetition, "error": false};
                }
                let element = delayElements[delayType][delay];
                if (isV6Result["error"] != null) {
                    element.setAttribute("class", "bg-danger");
                    let subelement = document.createElement("span");
                    subelement.textContent = 'err';
                    element.appendChild(subelement);
                } else if (isV6Result.isV6) {
                    element.setAttribute("class", "bg-success");
                    let subelement = document.createElement("span");
                    subelement.textContent = 'v6';
                    element.appendChild(subelement);
                } else {
                    element.setAttribute("class", "bg-warning");
                    let subelement = document.createElement("span");
                    subelement.textContent = 'v4';
                    element.appendChild(subelement);
                }
                await sleep(50);
            }
        }

        await sleep(500);
        observer.disconnect();
    } catch (error) {
        console.error('Error measuring network timings:', error);
    }

    for (const qtype of ['a', 'aaaa']) {
        for (const delay of configuredDelays) {
            const result = v6Infos[qtype][delay];
            result["responseTime"] = timingInfos[qtype][String(result['runUId'])];
            runInfo["delayResults"].push(result);
        }
    }

    const endDate = new Date();
    runInfo["timestampEnd"] = endDate.getTime();
}
