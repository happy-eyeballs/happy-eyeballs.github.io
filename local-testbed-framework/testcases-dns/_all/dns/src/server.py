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
    parser.add_argument("--listen6", help="enable dual stack listening", action='store_true')
    parser.add_argument("--port", help="UDP port the server will listen on")
    parser.add_argument("--local-ns-ip", nargs="+", help="IP address only used for DNS (not related to HE tests)")
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
        port: int,
        local_addr: str,
        ancdata
):
    question = DNSRecord.parse(packet)
    answer = resolver.resolve(question, queue, addr, port, local_addr)
    socket.sendmsg([answer.pack()], ancdata, 0, (addr, port))
    return


def main() -> None:
    parser = init_argparse()
    args = parser.parse_args()
    logging.basicConfig(level=logging.DEBUG)

    if not args.csv:
        logging.error("No log file location specified!")
        sys.exit(1)

    bind_address = '::'  # if args.listen6 else str(args.listen)
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
    with socket.socket(socket.AF_INET6, socket.SOCK_DGRAM) as s:
        s.bind((bind_address, bind_port))
        s.setsockopt(socket.IPPROTO_IP, socket.IP_PKTINFO, 1)
        s.setsockopt(socket.IPPROTO_IPV6, socket.IPV6_RECVPKTINFO, 1)

        with Manager() as manager:
            queue = manager.Queue()
            p = Process(target=log_to_file, args=(args.csv, queue))
            p.start()

            with Pool(processes=8) as pool:
                while not killer.kill:
                    try:
                        packet, ancdata, _, addr_info = s.recvmsg(4096, 4096)
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
                        logging.info(f'Connection from {addr}')
                        pool.apply_async(handle_request, (s, resolver, queue, packet, addr, port, dst_addr, ancdata))
                    except SystemExit:
                        pass

                logging.info('Stopping DNS server')

            p.close()
            p.join()

    del resolver


if __name__ == "__main__":
    main()
