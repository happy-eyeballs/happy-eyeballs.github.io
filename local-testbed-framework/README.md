# Local Testbed Framework

## Ansible setup

Given 2 hosts (client and server node) with 2 domain names in the same local domain reachable via ssh on the machine executing the ansible playbook, you can use the following comand to set these hosts up.
We suggest to use freshly installed systems (e.g., live boot systems).
`OS_NAME` can be for example debian and `OS_VERSION_NAME` bookworm.
The `ITERATIONS` variable decides how often the test run will be repeated.

```
ansible-playbook $DIR/setup.playbook.yml -i $DIR/inv \
  -e "local_domain='$LOCAL_DOMAIN'" \
  -e "client_node='$CLIENT_NODE'" \
  -e "server_node='$SERVER_NODE'" \
  -e "client_interface='$CLIENT_INTERFACE'" \
  -e "server_interface='$SERVER_INTERFACE'" \
  -e "os_name='$OS_NAME'" \
  -e "os_version_name='$OS_VERSION_NAME'" \
  -e "iterations='$ITERATIONS'" \
```

Afterward you can use the following command on the client node to start the test runs:

```
/opt/happyeyeballs/run.sh
```

After the test run finished the results are located at `/opt/happyeyeballs/artifacts` and the logs at `/opt/happyeyeballs/logs`.

Use `/opt/happyeyeballs/run.sh dns` to run the DNS tests.

## DNS setup for BIND

For some reason we could not make bind work without a public resolvable zone. Therefore, we one and redirected it to our internal name server IP addresses. Currently, this zone is called `internal.example.com`. Replace all instances with your own zone to also test bind. The zones name server should point to `10.0.2.1` and `fc00::2:1`. `dig internal.example.com @authoritative.of.example.com ns` should return something like:

```
;; AUTHORITY SECTION:
internal.example.com.	361 IN	NS	ns1.internal.example.com.

;; ADDITIONAL SECTION:
ns1.internal.example.com. 361	IN	A	10.0.2.1
ns1.internal.example.com. 361	IN	AAAA	fc00::2:1
```

## Repository Structure

- `ansible` contains the ansible playbook and roles to set up the client and server nodes
- `clients` contains the list of configured clients
- `clients-dns` contains the list of configured DNS resolver clients
- `local-run` contains the instructions, config, and scripts for running the test cases on the local machine (needed for Safari)
- `runner` contains the test runner written in Go
- `testcases` contains the list of configured Happy Eyeballs test cases
- `testcases-dns` contains the list of configured Happy Eyeballs DNS test cases
