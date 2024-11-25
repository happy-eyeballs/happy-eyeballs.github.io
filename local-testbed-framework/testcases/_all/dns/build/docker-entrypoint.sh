#!/bin/bash

python3 -u server.py --port 53 --zonefile dns.zone --csv queries.csv $@
