import re

from config import INCLUDE_REGEXP


def find_includes(file_path: str):
    # Если в файле есть прагма, Учесть ее
    with open(file_path) as file:
        for index, line in enumerate(file):
            found = re.fullmatch(INCLUDE_REGEXP, line)
            if found:
                yield index, found.group(0)
