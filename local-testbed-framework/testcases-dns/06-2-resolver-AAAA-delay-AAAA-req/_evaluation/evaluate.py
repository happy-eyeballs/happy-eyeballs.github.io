import sys
import dataclasses
from lib import *


def main():
    if len(sys.argv) != 4:
        print("Invalid command line argument count!", file=sys.stderr)
        exit(-1)

    pcap_json_path = sys.argv[1]
    ns_ipv4_addr = sys.argv[2]
    ns_ipv6_addr = sys.argv[3]
    test_case_iterations = load_packet_capture(pcap_json_path, True, ns_ipv4_addr, ns_ipv6_addr)

    outputs: list[Output] = []

    for test_case_iteration in test_case_iterations:
        attempted_address_families = [connection.address_family() for connection in test_case_iteration.connections]
        established_address_family = sorted([(connection.address_family(), connection.first_response_time()) for connection in test_case_iteration.connections if
                                        connection.has_response()], key=lambda x: x[1])[0][0]
        # time_until_first_connection = test_case_iteration.calculate_time_until_first_connection()

        first_attempted_ipv6_connection_index = attempted_address_families.index(IPFamily.IPV6) \
            if IPFamily.IPV6 in attempted_address_families else -1
        first_attempted_ipv4_connection_index = attempted_address_families.index(IPFamily.IPV4) \
            if IPFamily.IPV4 in attempted_address_families else -1

        happy_eyeballs_v1_connection_attempt_delay = None
        if first_attempted_ipv4_connection_index >= 0 and first_attempted_ipv6_connection_index >= 0:
            ipv6_connection_time = test_case_iteration.connections[first_attempted_ipv6_connection_index].time()
            ipv4_connection_time = test_case_iteration.connections[first_attempted_ipv4_connection_index].time()

            delay = ipv4_connection_time - ipv6_connection_time
            happy_eyeballs_v1_connection_attempt_delay = str(delay / datetime.timedelta(milliseconds=1))

        ipv6_syn_rtt = None
        if first_attempted_ipv6_connection_index >= 0:
            first_ipv6_connection = test_case_iteration.connections[first_attempted_ipv6_connection_index]
            response_time = first_ipv6_connection.first_response_time()
            if response_time is not None:
                rtt = response_time - first_ipv6_connection.time()
                ipv6_syn_rtt = str(rtt / datetime.timedelta(milliseconds=1))

        outputs.append(Output(
            iteration_variable=test_case_iteration.iteration_variable,
            measurement='Actual IPv6 SYN RTT',
            value=ipv6_syn_rtt,
        ))
        outputs.append(Output(
            iteration_variable=test_case_iteration.iteration_variable,
            measurement='Attempted Address Families',
            value=','.join(distinct_list(attempted_address_families)),
        ))
        outputs.append(Output(
            iteration_variable=test_case_iteration.iteration_variable,
            measurement='Established Address Families',
            value=established_address_family,
        ))
        # outputs.append(Output(
        #     iteration_variable=test_case_iteration.iteration_variable,
        #     measurement='Time Until First Response',
        #     value=time_until_first_connection
        # ))
        outputs.append(Output(
            iteration_variable=test_case_iteration.iteration_variable,
            measurement='Happy Eyeballs V1 Connection Attempt Delay',
            value=happy_eyeballs_v1_connection_attempt_delay
        ))

    # return outputs
    json.dump([dataclasses.asdict(output) for output in outputs], sys.stdout)


if __name__ == "__main__":
    main()
