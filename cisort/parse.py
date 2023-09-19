import re

#TODO(Pavel): Не забыть поставить точку
from settings import EXPRESSION


def get_includes(path) -> list:
    with open(path) as file:
        i = 0
        includes = []
        file_lines = file.readlines()
        file_lines.append('')
        for line in file_lines:
            line = line.rstrip()
            # print(i, line)
            if re.fullmatch(EXPRESSION, line):
                includes.append(
                    re.fullmatch(EXPRESSION, line).groups()[0]
                )
                print(i, file_lines.pop(i))
                i -= 2
            i += 1
    with open(path, 'w') as file:
        file.writelines(file_lines)

        return includes[:-1] if not includes[-1] else includes
