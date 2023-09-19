import re

# TODO(Pavel): Не забыть поставить точку
from settings import EXPRESSION


def extract_includes(path) -> list:
    with open(path) as file:
        i = 0
        includes = []
        file_lines = file.readlines()
        file_lines.append('')
        while i < len(file_lines):
            line = file_lines[i].rstrip()
            if re.fullmatch(EXPRESSION, line):
                includes.append(
                    re.fullmatch(EXPRESSION, line).groups()[0]
                )
                file_lines.pop(i)
                if i > 0:
                    i -= 1
            else:
                i += 1
    with open(path, 'w') as file:
        file.writelines(file_lines)

    return includes


def fix_lines(data):
    if len(data) == 0:
        return
    line = data[0]
    while line == '\n':
        data.pop(0)
        line = data[0]


def insert_includes(includes, path, flags):
    with open(path) as file:
        data = file.readlines()
        fix_lines(data)
        j = 0
        for i, key in enumerate(includes):
            if includes[key] and ('-c' in flags or '--comments' in flags):
                data.insert(j, f'// {key}\n')
                j += 1
            for include in includes[key]:
                data.insert(j, f'#include {include}\n')
                j += 1
            if includes[key]:
                data.insert(j, '\n')
                j += 1

    with open(path, 'w') as file:
        file.writelines(data)
