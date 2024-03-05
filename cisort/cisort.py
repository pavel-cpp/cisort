import logging
import os
import sys
from cisort.utils.loader import find_files

from config import (STARTUP_MESSAGE, configure_argument_parser,
                    configure_logging)


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


def cisort():
  configure_logging()
  arg_parser = configure_argument_parser()
  args = arg_parser.parse_args()
  if args.verbose:
    logging.info(STARTUP_MESSAGE)

  for files in find_files(args.files):
    for file in files:
        logging.info(f'Cisorting {file}...')
    insert_includes(sort_all(get_includes(file)), file)
    if not args.recursive:
        break
  print(f'{len(files)} files are cisorted!')


if __name__ == '__main__':
  cisort()
