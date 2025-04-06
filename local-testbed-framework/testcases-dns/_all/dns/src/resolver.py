from multiprocessing import Queue
from argparse import Namespace
from dnslib import DNSRecord, DNSQuestion, DNSLabel, RR, QTYPE, CLASS, RCODE, TXT

from logger import LogItem

import collections
import ipaddress
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
            local_addr
    ) -> DNSRecord:
        """Resolve the DNS request to a DNS response."""
        reply = request.reply()
        question: DNSQuestion = request.q

        orig_qname: DNSLabel = question.qname
        qname = DNSLabel(orig_qname.idna().lower())
        qclass: str = CLASS[question.qclass]
        qtype: str = QTYPE[question.qtype]
        local_addr = local_addr if ipaddress.ip_address(local_addr).ipv4_mapped is None else str(ipaddress.ip_address(local_addr).ipv4_mapped)

        log.put(LogItem(
            id=request.header.id,
            type="QUESTION",
            peer_addr=addr,
            peer_port=str(port),
            rr_name=qname.idna(),
            rr_class=qclass,
            rr_type=qtype,
        ))

        delay_ipv6 = self._delay_ipv6
        delay_ipv4 = self._delay_ipv4

        delay_label = None
        for label in qname.label:
            decoded_label: str = label.lower().decode()

            if decoded_label.startswith('delay_a-'):
                delay_ipv4 = int(decoded_label[8:])
                delay_label = decoded_label
            elif decoded_label.startswith('delay_aaaa-'):
                delay_ipv6 = int(decoded_label[11:])
                delay_label = decoded_label

        qname = DNSLabel(qname.label)

        reply.header.set_ad(False)
        reply.header.set_ra(False)
        first_label = qname.label[0].decode()
        delegation = False
        skip_glue = False
        new_ns_id = 'missing'
        # Do not "skip" delayed NS record delegation. Only provide dns delay info from he addresses
        if (qname.matchGlob('*.dns-delay.*') or qname.matchGlob('*.dns-delay-wg.*')) and (local_addr in self._local_ns_ips):
            if qname.matchGlob('*.dns-delay-wg.*'):
                skip_glue = True
            id_label = first_label
            if not id_label.startswith('id-'):
                reply.header.rcode = RCODE.REFUSED
                return reply
            if qname.label[1].lower().decode().startswith('delay_a'):
                qname = DNSLabel(qname.label[2:])
            else:
                qname = DNSLabel(qname.label[1:])
            qtype = 'NS'
            delegation = True
            # new_ns_id = random.randint(0, 1_000_000)
            new_ns_id = id_label.split('-')[1]
            reply.header.set_aa(False)
        else:
            if qtype == 'AAAA' and delay_ipv6 > 0:
                time.sleep(delay_ipv6 / 1000)
            if qtype == 'A' and delay_ipv4 > 0:
                time.sleep(delay_ipv4 / 1000)

        # Search for matching RRs
        for rr in self.get_records(qclass, qtype):
            if qname == rr.rname or qname.matchGlob(rr.rname):
                crr = self._correct_rr(rr, qname)

                if delegation:
                    first_label = crr.rdata.get_label().label[0].decode()
                    if first_label == 'ns1-id---' or first_label == 'ns2-id---':
                        if delay_label:
                            crr.rdata.set_label([f'{first_label[:7]}{new_ns_id}'.encode(), delay_label.encode(), *crr.rdata.get_label().label[1:]])
                        else:
                            crr.rdata.set_label([f'{first_label[:7]}{new_ns_id}'.encode(), *crr.rdata.get_label().label[1:]])
                    # logging.info(crr.rdata)
                    reply.add_auth(crr)
                else:
                    reply.add_answer(crr)

                log.put(LogItem(
                    id=request.header.id,
                    type="ANSWER",
                    peer_addr=addr,
                    peer_port=str(port),
                    rr_name=crr.rname.idna(),
                    rr_class=CLASS[crr.rclass],
                    rr_type=QTYPE[crr.rtype],
                    rr_value=crr.rdata,
                ))

                if qtype == 'NS' and not skip_glue:
                    # Search for glue records
                    for other_rr in (self.get_records('IN', 'A') + self.get_records('IN', 'AAAA')):
                        if crr.rdata.get_label() == other_rr.rname or crr.rdata.get_label().matchGlob(other_rr.rname):
                            other_crr = self._correct_rr(other_rr, crr.rdata.get_label())
                            reply.add_ar(other_crr)

        # No RRs found?
        if not reply.rr and not reply.auth:
            match = False
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
                    log.put(LogItem(
                        id=request.header.id,
                        type="NXDOMAIN",
                        peer_addr=addr,
                        peer_port=str(port),
                        rr_name=qname.idna(),
                        rr_class=qclass,
                        rr_type=qtype,
                    ))
                    reply.add_auth(self.find_record(qname, self.SOA))
                    reply.header.rcode = RCODE.NXDOMAIN
                else:
                    log.put(LogItem(
                        id=request.header.id,
                        type="REFUSED",
                        peer_addr=addr,
                        peer_port=str(port),
                        rr_name=qname.idna(),
                        rr_class=qclass,
                        rr_type=qtype,
                    ))
                    reply.header.rcode = RCODE.REFUSED

        return reply
