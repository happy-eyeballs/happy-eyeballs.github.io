import sys
import dataclasses
from lib import *


def main():
    if len(sys.argv) != 2:
        print("Invalid command line argument count!", file=sys.stderr)
        exit(-1)

    pcap_json_path = sys.argv[1]
    test_case_iterations = load_packet_capture(pcap_json_path, True)

    outputs: list[Output] = []

    for test_case_iteration in test_case_iterations:
        attempted_address_families = [connection.address_family() for connection in test_case_iteration.connections]
        established_address_families = [connection.address_family() for connection in test_case_iteration.connections if
                                        connection.contains_http_packet()]
        time_until_first_connection = test_case_iteration.calculate_time_until_first_connection()

        # the Happy Eyeballs v2 'Resolution Delay' is the time the client waits after the arrival of an A response for
        # the arrival of the AAAA response, before it tries an IPv4 connection establishment
        happy_eyeballs_v2_resolution_delay = None
        dns_a_answer_packets = [packet for packet in test_case_iteration.dns_packets if
                                packet.has_dns_a_record_answers()]
        ipv4_connection_attempts = [connection for connection in test_case_iteration.connections if
                                    connection.address_family() == IPFamily.IPV4]
        if len(dns_a_answer_packets) > 0 and len(ipv4_connection_attempts) > 0:
            delay = ipv4_connection_attempts[0].time() - dns_a_answer_packets[0].time()
            happy_eyeballs_v2_resolution_delay = str(delay / datetime.timedelta(milliseconds=1))

        outputs.append(Output(
            iteration_variable=test_case_iteration.iteration_variable,
            measurement='Attempted Address Families',
            value=','.join(distinct_list(attempted_address_families)),
        ))
        outputs.append(Output(
            iteration_variable=test_case_iteration.iteration_variable,
            measurement='Established Address Families',
            value=','.join(distinct_list(established_address_families)),
        ))
        outputs.append(Output(
            iteration_variable=test_case_iteration.iteration_variable,
            measurement='Time Until First Connection',
            value=time_until_first_connection
        ))
        outputs.append(Output(
            iteration_variable=test_case_iteration.iteration_variable,
            measurement='Happy Eyeballs V2 Resolution Delay',
            value=happy_eyeballs_v2_resolution_delay
        ))

    json.dump([dataclasses.asdict(output) for output in outputs], sys.stdout)


if __name__ == "__main__":
    main()
