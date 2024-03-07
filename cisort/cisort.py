import logging
import os
import sys

from config import (PRAGMA, STARTUP_MESSAGE, configure_argument_parser,
                    configure_logging)
from utils.files import find_files


#TODO: переработать
def insert_includes(includes_with_indices, path):
    includes_with_indices = sorted(includes_with_indices, key=lambda x: x[1])
    with open(path) as file:
        lines = file.readlines()
    start_index = lines[0] == PRAGMA
    # Возможно объединить два цикла
    for index, _ in includes_with_indices:
        lines.pop(index)
    # TODO: Добавить группировку (ЭТА ПОТОМ)
    for _, include in includes_with_indices:
        lines.insert(start_index, include)
    with open(path, 'w') as file:
        file.writelines(lines)


def cisort():
    configure_logging()
    arg_parser = configure_argument_parser()
    args = arg_parser.parse_args()
    if args.verbose:
        logging.info(STARTUP_MESSAGE)
    
    cnt = 0

    for file_paths in find_files(args.files):
        for file_path in file_paths:
            logging.info(f'Cisorting {file_path}...')

            cnt += 1
        if not args.recursive:
            break

    logging.info(f'{cnt} file{"s" if cnt > 1 else ""} are cisorted!')


if __name__ == '__main__':
    cisort()
