import os
import re
import sys


def get_includes(path):
    with open(path) as file:
        expr = r'^#include\s*([<\"].[^>\"]*[>\"])\n'
        index = end = 0
        includes = [[]]
        for line in file:
            end += 1
            if re.fullmatch(expr, line):
                includes[index].append(re.fullmatch(expr, line).groups()[0])
            elif len(includes[-1]) != 0:
                includes[index].append(
                    (end - len(includes[index]) - 1, end - 1))
                index += 1
                includes.append([])

        return includes[:-1] if not includes[-1] else includes


def sort_all(includes):
    for include in includes:
        coords = include.pop()
        include.sort()
        include.append(coords)
    return includes


def insert_includes(includes, path):
    with open(path) as file:
        data = file.readlines()
        for include in includes:
            for i in range(include[-1][0], include[-1][1]):
                data[i] = f'#include {include[i - include[-1][0]]}\n'
    with open(path, 'w') as file:
        file.writelines(data)


def get_files(directory: str = '.', files=None):
    if files is None:
        files = []
    for file in os.listdir(directory):
        path = f"{directory}/{file}"
        if os.path.isdir(path):
            get_files(path, files)
        elif file.split('.')[-1] in ('c', 'cpp', 'h', 'hpp'):
            print(f'[cisort] Add to sorting list {path}')
            files.append(path)
    return files


def cisort():
    if len(sys.argv) > 1:
        files = get_files(sys.argv[1])
    else:
        files = get_files()

    for file in files:
        print(f'[cisort] Sorting {file}')
        insert_includes(sort_all(get_includes(file)), file)


if __name__ == '__main__':
    cisort()
