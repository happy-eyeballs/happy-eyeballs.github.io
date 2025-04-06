
import * as main from './main.js';

document.getElementById("startTestBtn").addEventListener("click", main.measureHappyEyeballsV2);
document.getElementById("transmitResultsBtn").addEventListener("click", main.transmitResults);

let ipv2UserFormIds = ["startTestBtn", "transmitResultsBtn", "repetitions", "domainRandomization", "userInfo", "resolverInfo", "autoTransmit"];
main.setUserFormIds(ipv2UserFormIds);
main.setResultsPath("v2results");

document.addEventListener('DOMContentLoaded', main.setup);
