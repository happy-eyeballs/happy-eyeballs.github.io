from multiprocessing import Queue
from argparse import Namespace
from dnslib import DNSRecord, DNSQuestion, DNSLabel, RR, QTYPE, CLASS, RCODE, TXT

from logger import LogItem

import typing
import time


class Resolver:

    def __init__(self, zone: str, args: Namespace):
        """Initialize DNS resolver."""
        self._zone = RR.fromZone(zone)
        self._delay_ipv6 = 0 if args.delay_ipv6 is None else int(args.delay_ipv6)
        self._delay_ipv4 = 0 if args.delay_ipv4 is None else int(args.delay_ipv4)

        # Categorize RRs according to their QTYPE
        self._rr = {}
        for rr in self._zone:
            rtype = QTYPE[rr.rtype]
            try:
                self._rr[rtype].append(rr)
            except:
                self._rr[rtype] = [rr]

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
            port: int
    ) -> DNSRecord:
        """Resolve the DNS request to a DNS response."""
        reply = request.reply()
        question: DNSQuestion = request.q

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

        delay_ipv6 = self._delay_ipv6
        delay_ipv4 = self._delay_ipv4

        non_option_labels = []

        for label in qname.label:
            decoded_label: str = label.lower().decode()

            if decoded_label.startswith('delay_a-'):
                delay_ipv4 = int(decoded_label[8:])
                continue
            elif decoded_label.startswith('delay_aaaa-'):
                delay_ipv6 = int(decoded_label[11:])
                continue
            elif decoded_label.startswith('nonce-'):
                continue

            non_option_labels.append(label)

        qname = DNSLabel(non_option_labels)

        if qtype == 'AAAA' and delay_ipv6 > 0:
            time.sleep(delay_ipv6 / 1000)
        if qtype == 'A' and delay_ipv4 > 0:
            time.sleep(delay_ipv4 / 1000)

        # Search for matching RRs
        for rr in self.get_records(qclass, qtype):
            if qname == rr.rname:
                crr = self._correct_rr(rr, orig_qname)
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

                reply.add_answer(crr)

                if qtype == 'NS':
                    # Search for glue records
                    for other_rr in (self.get_records('IN', 'A') + self.get_records('IN', 'AAAA')):
                        if rr.rdata.get_label() == other_rr.rname:
                            reply.add_ar(other_rr)

        # No RRs found?
        if not reply.rr:
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
