import sys
import dataclasses
from lib import *


def main():
    if len(sys.argv) != 2:
        print("Invalid command line argument count!", file=sys.stderr)
        exit(-1)

    pcap_json_path = sys.argv[1]
    test_case = load_packet_capture(pcap_json_path, False)[0]

    destination_addresses = [connection.destination_ip_address() for connection in test_case.connections if
                             connection.address_family() != IPFamily.UNKNOWN]

    outputs: list[Output] = [
        Output(
            iteration_variable=0,
            measurement='Destination Addresses',
            value=','.join(destination_addresses),
        ),
    ]

    json.dump([dataclasses.asdict(output) for output in outputs], sys.stdout)


if __name__ == "__main__":
    main()
