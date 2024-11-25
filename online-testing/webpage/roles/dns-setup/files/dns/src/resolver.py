from multiprocessing import Queue
from argparse import Namespace
from dnslib import DNSRecord, DNSQuestion, DNSLabel, RR, QTYPE, CLASS, RCODE, TXT

from logger import LogItem

import collections
import ipaddress
import logging
import random
import typing
import time


class Resolver:

    def __init__(self, zone: str, args: Namespace):
        """Initialize DNS resolver."""
        self._zone = RR.fromZone(zone)
        self._delay_ipv6 = 0 if args.delay_ipv6 is None else int(args.delay_ipv6)
        self._delay_ipv4 = 0 if args.delay_ipv4 is None else int(args.delay_ipv4)

        self._local_ns_ips = args.local_ns_ip

        # Categorize RRs according to their QTYPE
        self._rr = collections.defaultdict(list)
        for rr in self._zone:
            rtype = QTYPE[rr.rtype]
            self._rr[rtype].append(rr)

        # Collect all existing domains and their corresponding SOA record
        domains: set[DNSLabel] = set([rr.rname for rr in self._zone])
        self._domains = [(domain, self.find_record(domain, self.SOA)) for domain in domains]
        # logging.info(f'domains available {self._domains}')

    @property
    def SOA(self) -> list[RR]:
        return self.get_records('IN', 'SOA')

    def get_records(self, rclass: str, rtype: str) -> list[RR]:
        if rclass == 'IN':
            try:
                return self._rr[rtype]
            except:
                pass
        return []

    def find_record(self, name: DNSLabel, rrs: list[RR]) -> typing.Optional[RR]:
        matches = list(filter(lambda record: name.idna().endswith(record.rname.idna()), rrs))
        if len(matches) > 0:
            # Do we have exact matches?
            exact = list(filter(lambda record: name == record.rname, matches))
            if len(exact) > 0:
                return exact[0]
            return matches[0]
        return None

    def _correct_rr(self, rr: RR, qname: DNSLabel) -> RR:
        return RR(
            qname,
            rr.rtype,
            rr.rclass,
            rr.ttl,
            rr.rdata
        )

    def resolve(
            self,
            request: DNSRecord,
            log: Queue,
            addr: str,
            port: int,
            local_addr: str
    ) -> DNSRecord:
        """Resolve the DNS request to a DNS response."""
        reply = request.reply()
        reply.header.set_ra(False)
        reply.header.set_ad(False)
        question: DNSQuestion = request.q
        local_addr = local_addr if ipaddress.ip_address(local_addr).ipv4_mapped is None else str(ipaddress.ip_address(local_addr).ipv4_mapped)

        orig_qname: DNSLabel = question.qname
        qname = DNSLabel(orig_qname.idna().lower())
        qclass: str = CLASS[question.qclass]
        qtype: str = QTYPE[question.qtype]

        log.put(LogItem(
            id=request.header.id,
            type="QUESTION",
            peer_addr=addr,
            peer_port=str(port),
            rr_name=qname.idna(),
            rr_class=qclass,
            rr_type=qtype,
        ))

        query_info = {
            'ns_ip': local_addr,
            'remote_ip': addr if ipaddress.ip_address(addr).ipv4_mapped is None else str(ipaddress.ip_address(addr).ipv4_mapped),
            'remote_port': port,
            'dns_query_id': request.header.id,
            'rr_name': qname.idna(),
            'rr_class': qclass,
            'rr_type': qtype,
            'answers': []
        }

        delay_ipv6 = self._delay_ipv6
        delay_ipv4 = self._delay_ipv4

        first_label = qname.label[0].decode()
        delay_label = None
        delay_v4 = False
        if first_label.startswith('v2delay_a-'):
            delay_label = first_label[10:]
            delay_v4 = True
        if first_label.startswith('v2delay_aaaa-'):
            delay_label = first_label[13:]

        if delay_label:
            query_id, delay = delay_label.split('_')
            query_info['id'] = query_id
            if delay_v4:
                delay_ipv4 = int(delay)
            else:
                delay_ipv6 = int(delay)
            query_info['delay_ms'] = int(delay)

        if first_label.startswith('id-'):
            query_info['id'] = first_label.split('-')[1]

        qname = DNSLabel(qname.label)

        if qtype == 'AAAA' and delay_ipv6 > 0:
            time.sleep(delay_ipv6 / 1000)
        if qtype == 'A' and delay_ipv4 > 0:
            time.sleep(delay_ipv4 / 1000)

        delegation = False
        sendglue = False
        new_ns_id = 'missing'
        # Do not "skip" delayed NS record delegation. Only provide dns delay info from he addresses
        if (qname.matchGlob('*.dns-delay-*.v1-rdns.example.com.') or qname.matchGlob('*.v6ns-only.v1-rdns.example.com.')) and (local_addr in self._local_ns_ips):
            id_label = first_label
            if not id_label.startswith('id-'):
                reply.header.rcode = RCODE.REFUSED
                return reply, None
            # qname = DNSLabel(qname.label[1:])
            qtype = 'NS'
            delegation = True
            # new_ns_id = random.randint(0, 1_000_000)
            new_ns_id = id_label.split('-')[1]
            if new_ns_id.startswith('wg'):
                sendglue = True
            reply.header.set_aa(False)

        query_info['delegation'] = delegation

        # Search for matching RRs
        for rr in self.get_records(qclass, qtype):
            if qname == rr.rname or qname.matchGlob(rr.rname):
                crr = self._correct_rr(rr, qname)

                log.put(LogItem(
                    id=request.header.id,
                    type="ANSWER" if not delegation else "AUTHORITY",
                    peer_addr=addr,
                    peer_port=str(port),
                    rr_name=crr.rname.idna(),
                    rr_class=CLASS[crr.rclass],
                    rr_type=QTYPE[crr.rtype],
                    rr_value=crr.rdata,
                ))

                if delegation:
                    first_label = crr.rdata.get_label().label[0].decode()
                    if first_label == 'ns1-id---' or first_label == 'ns2-id---':
                        crr.rdata.set_label([f'{first_label[:7]}{new_ns_id}'.encode(), *crr.rdata.get_label().label[1:]])
                    logging.info(crr.rdata)
                    reply.add_auth(crr)
                else:
                    reply.add_answer(crr)

                query_info['answers'].append({'rname': crr.rname.idna(), 'rtype': QTYPE[crr.rtype], 'rvalue': str(crr.rdata)})

                # skip these to better track resolvers
                if sendglue:
                    # Search for glue records
                    for other_rr in (self.get_records('IN', 'A') + self.get_records('IN', 'AAAA')):
                        if rr.rdata.get_label() == other_rr.rname or rr.rdata.get_label().matchGlob(other_rr.rname):
                            reply.add_ar(other_rr)

        # No RRs found?
        if not reply.rr and not reply.auth:
            match = False
            query_info = None
            for name, rr in self._domains:
                if name == qname:
                    log.put(LogItem(
                        id=request.header.id,
                        type="AUTHORITY",
                        peer_addr=addr,
                        peer_port=str(port),
                        rr_name=rr.rname.idna(),
                        rr_class=CLASS[rr.rclass],
                        rr_type=QTYPE[rr.rtype],
                        rr_value=rr.rdata,
                    ))
                    reply.add_auth(rr)
                    match = True
                    break

            # Did any records for the queried domain exist?
            if not match:
                # Is the queried domain part of our zone?
                if self.find_record(qname, self.SOA):
                    # log.put(LogItem(
                    #     id=request.header.id,
                    #     type="NXDOMAIN",
                    #     peer_addr=addr,
                    #     peer_port=str(port),
                    #     rr_name=qname.idna(),
                    #     rr_class=qclass,
                    #     rr_type=qtype,
                    # ))
                    reply.add_auth(self.find_record(qname, self.SOA))
                    reply.header.rcode = RCODE.NXDOMAIN
                    query_info = None
                else:
                    # log.put(LogItem(
                    #     id=request.header.id,
                    #     type="REFUSED",
                    #     peer_addr=addr,
                    #     peer_port=str(port),
                    #     rr_name=qname.idna(),
                    #     rr_class=qclass,
                    #     rr_type=qtype,
                    # ))
                    reply.header.rcode = RCODE.REFUSED
                    query_info = None

        return reply, query_info
