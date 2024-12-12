# Lazy Eye Inspection: Capturing the State of Happy Eyeballs Implementations

## Abstract

Happy Eyeballs (HE) started out by describing a mechanism that prefers IPv6 connections while ensuring a fast fallback to IPv4 when IPv6 fails. The IETF is currently working on the third version of HE. While the standards include recommendations for HE parameters choices, it is up to the client and OS to implement HE. In this paper we investigate the state of HE in various clients, particularly web browsers and recursive resolvers. We introduce a framework to analyze and measure client’s HE implementations and parameter choices. According to our evaluation, only Safari supports all HE features. Safari is also the only client implementation in our study that uses a dynamic IPv4 connection attempt delay, a resolution delay, and interlaces addresses. We further show that problems with the DNS A record lookup can even delay and interrupt the network connectivity despite a fully functional IPv6 setup with Chrome and Firefox. We publish our testbed measurement framework and a web-based tool to test HE properties on arbitrary browsers.

## Local Testbed Framework

- [Source Code](https://github.com/happy-eyeballs/happy-eyeballs.github.io/tree/main/local-testbed-framework)

## Online Test Tools

We will make our online test tool available to the public when double-blind restrictions are lifted. You can find a video of a testrun with our tool below.

- [Website Setup](https://github.com/happy-eyeballs/happy-eyeballs.github.io/tree/main/online-testing/webpage)
- [Open Resolver Scan](https://github.com/happy-eyeballs/happy-eyeballs.github.io/tree/main/online-testing/dnsscan)

### Safari Test Run

<a href="website-testrun.webm">Safari test run</a> showing its inconsistent HE behavior. In this example Safari uses an apparent connection attempt delay (CAD) of 250ms.

<video controls width="100%">
  <source src="website-testrun.webm" type="video/webm" />
</video>

## Overview of Browser Results


| Browser | Version | IPv6 Preferred | CAD | AAAA first<sup>1</sup> | RD Impl. | Num IPv4 Used | Num IPv6 Used | Addr. Selection | Local/Webtest Consistency |
|:--------|--------:|---------------:|----:|-----------:|---------:|--------------:|--------------:|----------------:|--------------:|
| Google Chrome  |   130.0 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">300ms</span> | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:red;color:white">No<sup>2</sup></span> | 1 | 1 | <span style="background-color:red;color:white">No</span> | <span style="background-color:green;color:white">Yes</span> |
| Chromium  |   130.0 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">300ms</span> | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:red;color:white">No<sup>2</sup></span> | 1 | 1 | <span style="background-color:red;color:white">No</span> | <span style="background-color:green;color:white">Yes</span> |
| Microsoft Edge  |   130.0 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">300ms</span> | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:red;color:white">No</span> | 1 | 1 | <span style="background-color:red;color:white">No</span> | <span style="background-color:green;color:white">Yes</span> |
| Mozilla Firefox  |   132.0 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">250ms</span> | <span style="background-color:red;color:white">No</span> | <span style="background-color:red;color:white">No</span> | 1 | 1 | <span style="background-color:red;color:white">No</span> | <span style="background-color:orange;color:white">Mostly<sup>3</sup></span> |
| Safari |   17.6 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">50ms - 2s</span> | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">Yes</span> | all | all | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:red;color:white">No<sup>4</sup></span> |
|<td colspan=9>**Mobile Browsers** (*No Addresses selection tests available with our website testing*)</td> |
| Mobile Safari |   17.6 & 18.1 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">50ms - 1s</span> | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">Yes</span> | - | - | - | - |
| Google Chrome Mobile |  127.0 & 130.0 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">300ms</span> | <span style="background-color:red;color:white">No</span> | <span style="background-color:red;color:white">No</span> | - | - | - | - |
| Mozilla Firefox Mobile |   125.0, 128.0, & 131.0 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">300ms</span> | <span style="background-color:red;color:white">No</span> | <span style="background-color:red;color:white">No</span> | - | - | - | - |
| <td colspan=9>**Command Line Tools** (*Wget does not perform any type of HE*)</td> |
| curl |   7.88.1 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:green;color:white">200ms</span> | <span style="background-color:gree;color:white">Yes</span> | <span style="background-color:red;color:white">No</span> | 1 | 1 | <span style="background-color:red;color:white">No</span> | - |
| wget |   1.21.3 | <span style="background-color:green;color:white">Yes</span> | <span style="background-color:red;color:white">-</span> | <span style="background-color:red;color:white">No</span> | <span style="background-color:red;color:white">No</span> | 0 | 1 | <span style="background-color:red;color:white">No</span> | - |


<sup>1</sup> May also be influenced by the operating system's stub resolver except for Chromium-based browsers which use their own stub resolver.
<sup>2</sup> Chromium and Chrome offer a feature flag to enable this feature. Possibly in future this will be enabled by default for all Chromium-based browsers.
<sup>3</sup> The observed multiple CAD values for Mozilla Firefox. Nevertheless, 250ms was the dominating value.
<sup>4</sup> Safari uses a dynamic approach. We could not determine scenarios or configurations which trigger a specific result.
