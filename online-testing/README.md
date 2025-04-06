# Online Test Setup

- we replaced all author specific IP addresses with documentation addresses (192.0.2.0/24 and 2001:db8::/64). To deploy this at your place replace all instaces of these documentation prefixes with your own. Make sure your host (host.example.com in `webpage/hosts`) can actually obtain all these addresses without issues
- we replaced the author specific domain name with `example.com` make sure you have the following sub zones with delegation to the corresponding IP addresses:
    - `v1` subdomain
        - 192.0.2.190
        - 192.0.2.191
        - 2001:db8::d:1
        - 2001:db8::d:2
    - `v1-rdns` subdomain
        - 192.0.2.190
        - 192.0.2.191
        - 2001:db8::d:1
        - 2001:db8::d:2
    - `v2` subdomain
        - 192.0.2.190
        - 192.0.2.191
        - 2001:db8::d:1
        - 2001:db8::d:2
- Replace all occurrences of `host.example.com` with the domain name of the host the webiste runs on
