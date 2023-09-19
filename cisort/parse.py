import re

from .settings import EXPRESSION


def get_includes(path) -> list:
    with open(path) as file:
        index = end = 0
        includes = [[]]
        file_lines = file.readlines()
        file_lines.append('')
        for line in file_lines:
            end += 1
            line = line.rstrip()
            if re.fullmatch(EXPRESSION, line):
                includes[index].append(
                    re.fullmatch(EXPRESSION, line).groups()[0]
                )
            elif len(includes[-1]) != 0:
                includes[index].append(
                    (end - len(includes[index]) - 1, end - 1)
                )
                index += 1
                includes.append([])

        return includes[:-1] if not includes[-1] else includes
