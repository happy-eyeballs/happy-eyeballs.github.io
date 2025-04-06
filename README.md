# Lazy Eye Inspection: Capturing the State of Happy Eyeballs Implementations

Paper ([Preprint](https://arxiv.org/pdf/2412.00263)) got accepted for publication at ACM IMC 2025.

## Authors
Patrick Sattler, Matthias Kirstein, Lars Wüstrich, Johannes Zirngibl, Georg Carle

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


<table style="overflow: scroll; display: block;">
  <thead>
    <tr>
      <th style="text-align: left">Browser</th>
      <th style="text-align: right">Version</th>
      <th style="text-align: right">IPv6 Preferred</th>
      <th style="text-align: right">CAD</th>
      <th style="text-align: right">AAAA first<sup style="vertical-align: super">1</sup></th>
      <th style="text-align: right">RD Impl.</th>
      <th style="text-align: right">Num IPv4 Used</th>
      <th style="text-align: right">Num IPv6 Used</th>
      <th style="text-align: right">Addr. Selection</th>
      <th style="text-align: right">Local/Webtest Consistency</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td style="text-align: left">Google Chrome</td>
      <td style="text-align: right">130.0</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">300ms</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No<sup style="vertical-align: super">2</sup></span></td>
      <td style="text-align: right">1</td>
      <td style="text-align: right">1</td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
    </tr>
    <tr>
      <td style="text-align: left">Chromium</td>
      <td style="text-align: right">130.0</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">300ms</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No<sup style="vertical-align: super">2</sup></span></td>
      <td style="text-align: right">1</td>
      <td style="text-align: right">1</td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
    </tr>
    <tr>
      <td style="text-align: left">Microsoft Edge</td>
      <td style="text-align: right">130.0</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">300ms</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right">1</td>
      <td style="text-align: right">1</td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
    </tr>
    <tr>
      <td style="text-align: left">Mozilla Firefox</td>
      <td style="text-align: right">132.0</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">250ms</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right">1</td>
      <td style="text-align: right">1</td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right"><span style="background-color:orange;color:white">Mostly<sup style="vertical-align: super">3</sup></span></td>
    </tr>
    <tr>
      <td style="text-align: left">Safari</td>
      <td style="text-align: right">17.6</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">50ms - 2s</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right">all</td>
      <td style="text-align: right">all</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No<sup style="vertical-align: super">4</sup></span></td>
    </tr>
    <tr>
      <td style="text-align: left;" colspan="10"><strong>Mobile Browsers</strong> (<em>No Addresses selection tests available with our website testing</em>)</td>
    </tr>
    <tr>
      <td style="text-align: left">Mobile Safari</td>
      <td style="text-align: right">17.6 &amp; 18.1</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">50ms - 1s</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
    </tr>
    <tr>
      <td style="text-align: left">Google Chrome Mobile</td>
      <td style="text-align: right">127.0 &amp; 130.0</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">300ms</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
    </tr>
    <tr>
      <td style="text-align: left">Mozilla Firefox Mobile</td>
      <td style="text-align: right">125.0, 128.0, &amp; 131.0</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">300ms</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
      <td style="text-align: right">-</td>
    </tr>
    <tr>
      <td style="text-align: left" colspan="10"><strong>Command Line Tools</strong> (<em>Wget does not perform any type of HE</em>)</td>
    </tr>
    <tr>
      <td style="text-align: left">curl</td>
      <td style="text-align: right">7.88.1</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:green;color:white">200ms</span></td>
      <td style="text-align: right"><span style="background-color:gree;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right">1</td>
      <td style="text-align: right">1</td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right">-</td>
    </tr>
    <tr>
      <td style="text-align: left">wget</td>
      <td style="text-align: right">1.21.3</td>
      <td style="text-align: right"><span style="background-color:green;color:white">Yes</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">-</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right">0</td>
      <td style="text-align: right">1</td>
      <td style="text-align: right"><span style="background-color:red;color:white">No</span></td>
      <td style="text-align: right">-</td>
    </tr>
  </tbody>
</table>

<sup style="vertical-align: super">1</sup> May also be influenced by the operating system's stub resolver except for Chromium-based browsers which use their own stub resolver.<br/>
<sup style="vertical-align: super">2</sup> Chromium and Chrome offer a feature flag to enable this feature. Possibly in future this will be enabled by default for all Chromium-based browsers.<br/>
<sup style="vertical-align: super">3</sup> The observed multiple CAD values for Mozilla Firefox. Nevertheless, 250ms was the dominating value.<br/>
<sup style="vertical-align: super">4</sup> Safari uses a dynamic approach. We could not determine scenarios or configurations which trigger a specific result.<br/>
