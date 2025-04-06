import datetime
from dataclasses import dataclass
from enum import StrEnum
import json


class IPFamily(StrEnum):
    UNKNOWN = 'Unknown'
    IPV4 = 'IPv4'
    IPV6 = 'IPv6'


class Packet:
    def __init__(self, layers: any):
        self._layers = layers

    def time(self) -> datetime:
        return datetime.datetime.fromtimestamp(float(self._layers['frame']['frame.time_epoch']))

    def ip_family(self) -> IPFamily:
        if 'ipv6' in self._layers:
            return IPFamily.IPV6

        if 'ip' in self._layers:
            return IPFamily.IPV4

        return IPFamily.UNKNOWN

    def is_connection_to(self, address) -> str:
        if self.ip_family() == IPFamily.IPV4:
            return address == self._layers['ip']['ip.dst'] or address == self._layers['ip']['ip.src']

        if self.ip_family() == IPFamily.IPV6:
            return address == self._layers['ipv6']['ipv6.dst'] or address == self._layers['ipv6']['ipv6.src']

    def destination_ip_address(self) -> str:
        if self.ip_family() == IPFamily.IPV4:
            return self._layers['ip']['ip.dst']

        if self.ip_family() == IPFamily.IPV6:
            return self._layers['ipv6']['ipv6.dst']

        return ''

    def is_tcp_packet(self) -> bool:
        return 'tcp' in self._layers

    def has_tcp_syn_flag(self) -> bool:
        return self.is_tcp_packet() and self._layers['tcp']['tcp.flags_tree']['tcp.flags.syn'] == '1'

    def has_tcp_ack_flag(self) -> bool:
        return self.is_tcp_packet() and self._layers['tcp']['tcp.flags_tree']['tcp.flags.ack'] == '1'

    def is_http_packet(self) -> bool:
        return 'http' in self._layers

    def get_http_url(self) -> str:
        if not self.is_http_packet():
            return ''

        if 'http.request.full_uri' in self._layers['http']:
            return self._layers['http']['http.request.full_uri']

        if 'http.response_for.uri' in self._layers['http']:
            return self._layers['http']['http.response_for.uri']

        return ''

    def get_non_dns_server_port(self) -> str:
        port1 = self._layers['udp']['udp.srcport']
        port2 = self._layers['udp']['udp.dstport']
        return port2 if port1 == '53' else port1

    def is_udp_packet(self) -> bool:
        return 'udp' in self._layers

    def get_udp_destination_port(self) -> int:
        return int(self._layers['udp']['udp.dstport'])

    def get_udp_data(self) -> str:
        return bytearray.fromhex(self._layers['udp']['udp.payload'].replace(':', '')).decode()

    def is_dns_packet(self) -> bool:
        return 'dns' in self._layers

    def is_dns_request_packet(self) -> bool:
        return self.is_dns_packet() and self.is_udp_packet() and self._layers['udp']['udp.dstport'] == '53'

    def has_dns_answers(self) -> bool:
        return (self.is_dns_packet() and 'dns.count.answers' in self._layers['dns'] and
                self._layers['dns']['dns.count.answers'] != '0')

    def has_dns_a_record_answers(self) -> bool:
        return (self.has_dns_answers() and
                len([rr for rr in self._layers['dns']['Answers'].keys() if 'type A,' in rr]) > 0)


class DNSConnection:
    def __init__(self, packets: list[Packet]):
        assert len(packets) > 0
        self.packets = packets

    def contains_request(self) -> bool:
        return self.packets[0].is_dns_request_packet()

    def has_response(self) -> bool:
        return len([p for p in self.packets if p.has_dns_answers()]) > 0

    def address_family(self) -> IPFamily:
        return self.packets[0].ip_family()

    def destination_ip_address(self) -> str:
        return self.packets[0].destination_ip_address()

    def contains_http_packet(self) -> bool:
        return len([packet for packet in self.packets if
                    packet.is_http_packet() and not packet.get_http_url().endswith('/favicon.ico')]) > 0

    def time(self) -> datetime.datetime:
        return self.packets[0].time()

    def first_response_time(self) -> datetime.datetime | None:
        addresses = [packet.destination_ip_address() for packet in self.packets]

        destination_address = addresses[0]
        source_address = [address for address in addresses if address != destination_address]
        if len(source_address) == 0:
            return None

        first_response_packet_index = addresses.index(source_address[0])
        return self.packets[first_response_packet_index].time()


class TestCaseIteration:
    def __init__(self, iteration_variable: int, packets_in_iteration: list[Packet], ns_ipv4_addr, ns_ipv6_addr):
        self.iteration_variable = iteration_variable

        self.connections = extract_dns_connections(packets_in_iteration, ns_ipv4_addr, ns_ipv6_addr)
        self.dns_packets = [packet for packet in packets_in_iteration if packet.is_dns_packet()]

    def calculate_time_until_first_connection(self) -> str | None:
        if len(self.connections) == 0:
            return None

        # retrieve all DNS packets sent by the client to the server
        dns_request_packets = [dns_packet for dns_packet in self.connections if dns_packet.is_dns_request_packet()]
        if len(dns_request_packets) == 0:
            return None

        time_until_first_request = self.connections[0].time() - dns_request_packets[0].time()
        return str(time_until_first_request / datetime.timedelta(milliseconds=1))


def extract_dns_connections(packets: list[Packet], ns_ipv4_addr, ns_ipv6_addr) -> list[DNSConnection]:
    # filter for tcp packets
    packets = [packet for packet in packets if packet.is_dns_packet() and (packet.is_connection_to(ns_ipv4_addr) or packet.is_connection_to(ns_ipv6_addr))]

    # map each packet to their connection
    packets_per_connection = dict()

    for packet in packets:
        # a connection is uniquely identified by the client-side IP address and client-side port, as the server-side
        # IP address is implied by the client's IP address and the server-side port is always the same for HTTP
        connection_id = f'{packet.ip_family()}|{packet.get_non_dns_server_port()}'

        if connection_id in packets_per_connection:
            packets_per_connection[connection_id].append(packet)
        else:
            packets_per_connection[connection_id] = [packet]

    connections = [DNSConnection(packets) for packets in packets_per_connection.values() if len(packets) > 0]

    # filter for connection captures that contain a handshake, because otherwise, the connection was already established
    # before this test case started and are therefore irrelevant for this test case
    connections = [connection for connection in connections if connection.contains_request()]

    return connections


def extract_test_case_iterations(captured_packets: list[Packet], ns_ipv4_addr, ns_ipv6_addr) -> list[TestCaseIteration]:
    # each test case iteration is separated by a UDP packet to port 44444
    separator_packets = [packet for packet in captured_packets if packet.is_udp_packet()
                         and packet.get_udp_destination_port() == 44444]

    assert len(separator_packets) > 0

    # skip all captured packets until first separator packet
    captured_packets = captured_packets[captured_packets.index(separator_packets[0]):]

    # split the captured packets at the separator packet indices
    iterations: list[TestCaseIteration] = []

    iteration_start_indices = [captured_packets.index(separator_packet) for separator_packet in separator_packets]
    iteration_end_indices = iteration_start_indices[1:] + [len(captured_packets)]

    for (start_index, end_index) in zip(iteration_start_indices, iteration_end_indices):
        packets_in_iteration = captured_packets[start_index:end_index]

        current_separator_packet = packets_in_iteration[0]
        iteration_variable = int(current_separator_packet.get_udp_data())  # variable is sent as udp payload

        iterations.append(TestCaseIteration(iteration_variable, packets_in_iteration, ns_ipv4_addr, ns_ipv6_addr))

    return iterations


def load_packet_capture(path: str, separator_packets: bool, ns_ipv4_addr, ns_ipv6_addr) -> list[TestCaseIteration]:
    capture_file = open(path)
    capture = json.load(capture_file)
    capture_file.close()

    captured_packets = [Packet(packet['_source']['layers']) for packet in capture]

    if separator_packets:
        return extract_test_case_iterations(captured_packets, ns_ipv4_addr, ns_ipv6_addr)
    else:
        return [TestCaseIteration(0, captured_packets, ns_ipv4_addr, ns_ipv6_addr)]


@dataclass
class Output:
    iteration_variable: int
    measurement: str
    value: str


def distinct_list(list_with_duplicates: list) -> list:
    index = set()
    return [element for element in list_with_duplicates if not (element in index or index.add(element))]
