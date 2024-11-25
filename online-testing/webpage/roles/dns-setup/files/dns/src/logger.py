from csv import DictWriter
from dataclasses import dataclass, fields, field, asdict
from multiprocessing import Queue

import time


def current_time() -> int:
    return time.time_ns()


@dataclass
class LogItem:
    id: int = 0
    type: str = ''
    peer_addr: str = ''
    peer_port: str = ''
    rr_name: str = ''
    rr_class: str = ''
    rr_type: str = ''
    rr_value: str = ''
    timestamp: int = field(default_factory=current_time)


def log_to_file(csv_file: str, q: Queue):
    with open(csv_file, "w", newline="") as file:
        fieldnames = [field.name for field in fields(LogItem)]

        writer = DictWriter(file, fieldnames=fieldnames, dialect="unix")
        writer.writeheader()
        file.flush()

        while True:
            item: LogItem = q.get()
            if item is None:
                break
            writer.writerow(asdict(item))
            file.flush()
