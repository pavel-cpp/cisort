import re

from config import INCLUDE_REGEXP


def find_includes(file_path):
    with open(file_path) as file:
        return {
            include: index
            for index, include in enumerate(
                re.findall(INCLUDE_REGEXP, file.read())
            )
        }
