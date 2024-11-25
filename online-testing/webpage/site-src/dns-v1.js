
import * as main from './main.js';

document.getElementById("startTestBtn").addEventListener("click", main.measureHappyEyeballsDNS);
document.getElementById("transmitResultsBtn").addEventListener("click", main.transmitResults);

document.addEventListener('DOMContentLoaded', main.setup);

let dnsv1UserFormIds = ["startTestBtn", "transmitResultsBtn", "repetitions", "runRandomization", "resolverInfo", "autoTransmit"];
main.setUserFormIds(dnsv1UserFormIds);
main.setResultsPath("dnsresults");
