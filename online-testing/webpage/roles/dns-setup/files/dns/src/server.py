from multiprocessing import Queue, Pool, Process, Manager
from resolver import Resolver
from dnslib import DNSRecord
from logger import log_to_file

import dnslib.dns

import datetime
import json
import os
import textwrap
import signal
import socket
import sys
import argparse
import logging
import multiprocessing
import requests
import ipaddress


DNSDATA_FILE = '{date}-dns-results.jsonl'
OUTPUT_DIR = None

# https://stackoverflow.com/q/49417041
class Killer:
    kill = False

    def __init__(self):
        signal.signal(signal.SIGINT, self.terminate)
        signal.signal(signal.SIGTERM, self.terminate)

    def terminate(self, *args):
        self.kill = True
        raise SystemExit


def init_argparse() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        usage="%(prog)s [OPTIONS]",
        description="Custom Python-based DNS server using a single DNS zonefile."
    )
    # parser.add_argument("--listen", help="IPv4 address the server will listen on")
    parser.add_argument("--listen6", help="enable dual stack listening", action='store_true')
    parser.add_argument("--port", help="UDP port the server will listen on")
    parser.add_argument("--zonefile", help="path to the DNS zonefile")
    parser.add_argument("--local-ns-ip", nargs="+", help="IP address only used for DNS (not related to HE tests)")
    parser.add_argument("--csv", help="location where to store the csv file containing metadata for all queries")
    parser.add_argument("--output-dir", help="host where to report v2 requests")
    parser.add_argument("--delay-ipv6", help="amount of time (ms) to delay a reply to an AAAA query")
    parser.add_argument("--delay-ipv4", help="amount of time (ms) to delay a reply to an A query")
    parser.add_argument("--v6delay-prefix", help="The prefix where tc delays are configured")
    return parser


def handle_request(
        s: socket.socket,
        resolver: Resolver,
        queue: Queue,
        packet: bytes,
        addr: str,
        port: int,
        local_addr: str,
        ancdata,
        cur_date: datetime.datetime,
        v6delay_prefix: ipaddress.IPv6Network
):
    try:
        question = DNSRecord.parse(packet)
    except dnslib.dns.DNSError as e:
        logging.exception(e)
        return

    if question.header.get_qr() != 0:
        return
    answer, query_data = resolver.resolve(question, queue, addr, port, local_addr)
    s.sendmsg([answer.pack()], ancdata, 0, (addr, port))
    if query_data:
        query_data['request_time'] = cur_date.timestamp()
        local_ip = ipaddress.ip_address(local_addr)
        if local_ip.version == 6 and local_ip in v6delay_prefix:
            query_data['delay_ms'] = int(local_ip.exploded.split(':')[-1])
            query_data['request_arrival_time'] = query_data['request_time']
            query_data['request_time'] = query_data['request_time'] - query_data['delay_ms'] / 1000
        elif 'delay_ms' not in query_data:
            query_data['delay_ms'] = 0

        write_data(query_data)
    return


def write_data(data):
    date = datetime.datetime.now()
    date_str = date.strftime('%Y-%m-%d')
    month_dir = date.strftime('%Y/%m')
    data_filepath = os.path.join(OUTPUT_DIR, month_dir, DNSDATA_FILE.format(date=date_str))
    os.makedirs(os.path.join(OUTPUT_DIR, month_dir), exist_ok=True)

    with open(data_filepath, 'a') as f:
        json.dump(data, f)
        f.write('\n')


def main() -> None:
    parser = init_argparse()
    args = parser.parse_args()
    logging.basicConfig(level=logging.DEBUG)
    logging.getLogger("urllib3").setLevel(logging.WARNING)

    multiprocessing.log_to_stderr(level=logging.INFO)

    if not args.csv:
        logging.error("No log file location specified!")
        sys.exit(1)

    global OUTPUT_DIR
    OUTPUT_DIR = args.output_dir

    bind_address = '::'
    bind_port = 53 if not args.port else int(args.port)

    # Allow graceful termination
    killer = Killer()

    # Load zone file
    zone = ''
    try:
        with open(args.zonefile, 'r') as f:
            for line in f:
                if line.startswith('#'):
                    continue
                zone += line
    except:
        logging.error('No DNS zone file found!')
        sys.exit(1)

    resolver = Resolver(textwrap.dedent(zone), args)

    v6delay_prefix = ipaddress.ip_network(args.v6delay_prefix)

    # Listen for incoming connections
    logging.info(f'Starting DNS server (listening on {bind_address} port {bind_port})')
    logging.info(f'local ns ips {args.local_ns_ip}')
    with socket.socket(socket.AF_INET6, socket.SOCK_DGRAM) as s:
        s.bind((bind_address, bind_port))
        s.setsockopt(socket.IPPROTO_IP, socket.IP_PKTINFO, 1)
        s.setsockopt(socket.IPPROTO_IPV6, socket.IPV6_RECVPKTINFO, 1)

        with Manager() as manager:
            queue = manager.Queue()
            p = Process(target=log_to_file, args=(args.csv, queue))
            p.start()

            with Pool(processes=150) as pool:
                while not killer.kill:
                    try:
                        packet, ancdata, _, addr_info = s.recvmsg(4096, 4096)
                        cur_date = datetime.datetime.now()
                        dst_addr = None
                        # logging.info(f'{ancdata}')
                        for cmsg_level, cmsg_type, cmsg_data in ancdata:
                            if cmsg_level == socket.IPPROTO_IPV6 and cmsg_type == socket.IPV6_PKTINFO:
                                dst_addr = socket.inet_ntop(socket.AF_INET6, cmsg_data[:16])
                                break
                            if cmsg_level == socket.IPPROTO_IP and cmsg_type == socket.IP_PKTINFO:
                                dst_addr = socket.inet_ntop(socket.AF_INET, cmsg_data[4:8])
                                break
                        addr = addr_info[0]
                        port = addr_info[1]
                        logging.info(f'Connection from {addr} towards {dst_addr}')
                        if dst_addr == '2001:db8::4e:ffff':
                            continue
                        pool.apply_async(handle_request, (s, resolver, queue, packet, addr, port, dst_addr, ancdata, cur_date, v6delay_prefix))
                        # handle_request(s, resolver, queue, packet, addr, port, dst_addr, ancdata, cur_date, v6delay_prefix)
                    except SystemExit:
                        pass

                logging.info('Stopping DNS server')

            p.close()
            p.join()

    del resolver


if __name__ == "__main__":
    main()
