
import * as main from './main.js';

document.getElementById("startTestBtn").addEventListener("click", main.measureHappyEyeballs);
document.getElementById("transmitResultsBtn").addEventListener("click", main.transmitResults);

document.addEventListener('DOMContentLoaded', main.setup);

let ipv1UserFormIds = ["startTestBtn", "transmitResultsBtn", "repetitions", "domainRandomization", "userInfo", "autoTransmit"];
main.setUserFormIds(ipv1UserFormIds);
main.setResultsPath("results");

