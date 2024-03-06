import logging
import os
import sys
from utils.files import find_files

from config import (STARTUP_MESSAGE, configure_argument_parser,
                    configure_logging)


def sort_all(includes: dict[str, int]):
    return sorted(includes.items(), key=lambda include: include[0])


# def insert_includes(includes, path):
#   with open(path) as file:
#     data = file.readlines()
#     for include in includes:
#       for i in range(include[-1][0], include[-1][1]):
#         data[i] = f'#include {include[i - include[-1][0]]}\n'
#   with open(path, 'w') as file:
#     file.writelines(data)

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
