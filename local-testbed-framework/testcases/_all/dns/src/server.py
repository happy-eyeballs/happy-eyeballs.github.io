from multiprocessing import Queue, Pool, Process, Manager
from resolver import Resolver
from dnslib import DNSRecord
from logger import log_to_file

import textwrap
import signal
import socket
import sys
import argparse
import logging


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
    parser.add_argument("--listen", help="IPv4 address the server will listen on")
    parser.add_argument("--port", help="UDP port the server will listen on")
    parser.add_argument("--zonefile", help="path to the DNS zonefile")
    parser.add_argument("--csv", help="location where to store the csv file containing metadata for all queries")
    parser.add_argument("--delay-ipv6", help="amount of time (ms) to delay a reply to an AAAA query")
    parser.add_argument("--delay-ipv4", help="amount of time (ms) to delay a reply to an A query")
    return parser


def handle_request(
        socket: socket.socket,
        resolver: Resolver,
        queue: Queue,
        packet: bytes,
        addr: str,
        port: int
):
    question = DNSRecord.parse(packet)
    answer = resolver.resolve(question, queue, addr, port)
    socket.sendto(answer.pack(), (addr, port))
    return


def main() -> None:
    parser = init_argparse()
    args = parser.parse_args()
    logging.basicConfig(level=logging.DEBUG)

    if not args.csv:
        logging.error("No log file location specified!")
        sys.exit(1)

    bind_address = '0.0.0.0' if not args.listen else str(args.listen)
    bind_port = 53 if not args.port else int(args.port)

    # Allow graceful termination
    killer = Killer()

    # Load zone file
    zone = ''
    try:
        with open(args.zonefile, 'r') as f:
            zone = '\n'.join(f.readlines())
    except:
        logging.error('No DNS zone file found!')
        sys.exit(1)

    resolver = Resolver(textwrap.dedent(zone), args)

    # Listen for incoming connections
    logging.info(f'Starting DNS server (listening on {bind_address} port {bind_port})')
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
        s.bind((bind_address, bind_port))

        with Manager() as manager:
            queue = manager.Queue()
            p = Process(target=log_to_file, args=(args.csv, queue))
            p.start()

            with Pool(processes=8) as pool:
                while not killer.kill:
                    try:
                        packet, (addr, port) = s.recvfrom(4096)
                        logging.info(f'Connection from {addr}')
                        pool.apply_async(handle_request, (s, resolver, queue, packet, addr, port))
                    except SystemExit:
                        pass

                logging.info('Stopping DNS server')

            p.close()
            p.join()

    del resolver


if __name__ == "__main__":
    main()
