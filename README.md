# Lazy Eye Inspection: Capturing the State of Happy Eyeballs Implementations

## Abstract

Happy Eyeballs (HE) started out by describing a mechanism that prefers IPv6 connections while ensuring a fast fallback to IPv4 when IPv6 fails. The IETF is currently working on the third version of HE. While the standards include recommendations for HE parameters choices, it is up to the client and OS to implement HE. In this paper we investigate the state of HE in various clients, particularly web browsers and recursive resolvers. We introduce a framework to analyze and measure client’s HE implementations and parameter choices. According to our evaluation, only Safari supports all HE features. Safari is also the only client implementation in our study that uses a dynamic IPv4 connection attempt delay, a resolution delay, and interlaces addresses. We further show that problems with the DNS A record lookup can even delay and interrupt the network connectivity despite a fully functional IPv6 setup with Chrome and Firefox. We publish our testbed measurement framework and a web-based tool to test HE properties on arbitrary browsers.

## Local Testbed Framework

- [Source Code](https://github.com/happy-eyeballs/happy-eyeballs.github.io/tree/main/local-testbed-framework)

## Online Test Tools

- [Website Setup](https://github.com/happy-eyeballs/happy-eyeballs.github.io/tree/main/online-testing/webpage)
- [Open Resolver Scan](https://github.com/happy-eyeballs/happy-eyeballs.github.io/tree/main/online-testing/dnsscan)

## Overview of Browser Results

- TBD
