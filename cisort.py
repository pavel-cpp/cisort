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


def get_files(directory: str = '.',
              flags: list | None = None,
              files: list | None = None) -> list:
    if flags is None:
        flags = []
    if files is None:
        files = []

    for file in os.listdir(directory):
        path = f"{directory}/{file}"
        if os.path.isdir(path) and '-r' in flags:
            get_files(path, flags, files)
        elif file.split('.')[-1] in ('c', 'cpp', 'h', 'hpp'):
            if '-ls' in flags:
                print(f'Add to cisorting list {path}')
            files.append(path)
    return files


def is_correct_flags(flags: list) -> bool:
    available_flags = ('-r', '-ls', '-h', '--help')
    for flag in flags:
        if flag not in available_flags:
            print(
                f'Flag "{flag}" is incorrect!\n'
                f'Try:\n\tcisort --help'
            )
            return False
    return True


def cisort():
    args = sys.argv[1:]

    if '-h' in args or '--help' in args:
        print(
            'Using:\n\n'
            '\tcisort [flags] [path]\n\n'
            'Flags:\n'
            '\t-r - recursive searching C/C++ files\n'
            '\t-ls - show info about sorted files\n'
            '\t-h --help - to get help'
        )
        return

    if sys.argv[-1][0] == '-' or not args:
        print('Start cisearching...')
        files = get_files(flags=args)
    elif is_correct_flags(args[:-1]):
        print('Start cisearching...')
        files = get_files(directory=args[-1], flags=args[:-1])
    else:
        return

    print('Start cisorting...')
    for file in files:
        if '-ls' in args:
            print(f'Cisorting {file}')
        insert_includes(sort_all(get_includes(file)), file)
    print(f'{len(files)} files are cisorted!')

if __name__ == '__main__':
    cisort()
