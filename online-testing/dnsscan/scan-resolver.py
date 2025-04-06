#!/bin/env python

import argparse
import fileinput
import functools
import json
import math
import multiprocessing.pool
import os.path
import random
import sys
import threading
import traceback

import dns.rdatatype
import dns.resolver


write_lock = threading.Lock()
RECORD_TYPE: dns.rdatatype.RdataType = dns.rdatatype.AAAA
WITH_GLUE = False
OUTFILE = sys.stdout

class ResolveException(Exception):
    def __init__(self, domain, msg):
        self.domain = domain
        self.msg = msg
        super(ResolveException, self).__init__(msg)


def write_output(resolver_address, delay, run_id, exc=None):
    with write_lock:
        print(f"{resolver_address},{delay},{run_id},{exc.msg if exc else ''}",
              file=OUTFILE,
              flush=True)
        if exc:
            print(f"{resolver_address},{delay},{run_id},{exc.domain}: {exc.msg}",
                  file=sys.stderr)


def get_delay_domain(basezone, delay, run_id):
    return f'id-{run_id}.dns-delay-{delay}.{basezone}'


def get_v6only_domain(basezone, run_id):
    return f'id-{run_id}.v6ns-only.{basezone}'


def get_run_id():
    randmin = 0
    randmax = 100_000
    return math.floor(random.random() * (randmax - randmin)) + randmin


def perform_requests(basezone, delays, resolver_address):
    resolver_address = resolver_address.strip()
    resolver = dns.resolver.Resolver()
    resolver.timeout = 7
    resolver.lifetime = 7
    resolver.nameservers = [resolver_address]

    try:
        perform_dns_query(get_v6only_domain(basezone, get_run_id()), resolver)
    except ResolveException as e:
        write_output(resolver_address, 0, 'v6only', e)

    for delay in delays:
        run_id = get_run_id()
        if WITH_GLUE:
            run_id = f'wg{run_id}'
        domain = get_delay_domain(basezone, delay, run_id)
        try:
            perform_dns_query(domain, resolver)
        except ResolveException as e:
            write_output(resolver_address, delay, run_id, e)
        else:
            write_output(resolver_address, delay, run_id)


def perform_dns_query(domain, resolver: dns.resolver.Resolver):
    # id-${randid}.dns-delay-${delay}.v1-dns.example.com
    try:
        resp = resolver.resolve(domain, RECORD_TYPE)
    except dns.resolver.Timeout:
        error = [domain, "Timeout"]
    except dns.resolver.NoNameservers:
        error = [domain, "No Nameserver could resolve this domain"]
    except dns.resolver.NoAnswer:
        error = [domain, "No answer (nodata)"]
    except dns.resolver.NXDOMAIN:
        error = [domain, "Query returned NXDOMAIN"]
    else:
        rrs = None
        for rr in resp.response.answer:
            if rr.rdtype == RECORD_TYPE and str(rr.name).strip('.') == domain:
                rrs = rr
                break
        if rrs:
            ip = None
            for rr in rrs:
                ip = str(rr.address)
            if RECORD_TYPE == dns.rdatatype.AAAA and ip != '2001:db8::4f:0':
                error = [domain, 'Incorrect resolution']
            elif RECORD_TYPE == dns.rdatatype.A and ip != '192.0.2.189':
                error = [domain, 'Incorrect resolution']
            else:
                return
        else:
            error = [domain, "No rr in answer"]

    raise ResolveException(*error)


def main():
    parser = argparse.ArgumentParser(description='Scan v1 HE resolver targets')
    parser.add_argument('resolver_file', type=str,
                        help='Input file containing a list of resolver IP addresses')
    parser.add_argument('delays_file', type=str,
                        help='Input file containing a list of delays to test')
    parser.add_argument('basezone', type=str,
                        help='The DNS zone where the test i deployed')
    parser.add_argument('--outfile', type=str,
                        help='Output file location')
    parser.add_argument('--record-type', type=str,
                        help='The record type to request', default='AAAA')
    parser.add_argument('--with-glue', action='store_true',
                        help='The record type to request')

    args = parser.parse_args()

    if args.resolver_file == '-':
        inf = fileinput.input()
    else:
        inf = open(args.resolver_file, "r")

    if os.path.exists(args.outfile):
        print('Outfile already exists', file=sys.stderr)
        sys.exit(1)

    delays = []
    with open(args.delays_file, "r") as f:
        for line in f:
            delays.append(line.strip())

    global RECORD_TYPE, OUTFILE, WITH_GLUE
    RECORD_TYPE = dns.rdatatype.RdataType[args.record_type]
    OUTFILE = open(args.outfile, 'w')
    WITH_GLUE = args.with_glue
    basezone = args.basezone
    pool = multiprocessing.pool.ThreadPool(processes=10)
    result = pool.map_async(functools.partial(perform_requests, basezone, delays), inf, chunksize=1)
    try:
        while result.wait():
            pass
    except StopIteration:
        print('finished', file=sys.stderr)
    pool.close()
    # for line in inf:
    #     line = line.strip()
    #     if not line:
    #         continue
    #     perform_requests(basezone, delays, line)
    if args.resolver_file != '-':
        inf.close()
    OUTFILE.close()


if __name__ == "__main__":
    main()
