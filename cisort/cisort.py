import logging
import os
import sys

from config import (
    configure_argument_parser,
    configure_logging,
    STARTUP_MESSAGE,
)


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


def get_files():
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


def cisort():
  configure_logging()
  arg_parser = configure_argument_parser()
  args = arg_parser.parse_args()
  if args.verbose:
    logging.info(STARTUP_MESSAGE)

  files = get_files(flags=args)

  print('Start cisearching...')
  files = get_files(directory=args[-1], flags=args[:-1])
  print('Start cisorting...')
  for file in files:
    if '-ls' in args:
      print(f'Cisorting {file}')
    insert_includes(sort_all(get_includes(file)), file)
  print(f'{len(files)} files are cisorted!')


if __name__ == '__main__':
  cisort()
