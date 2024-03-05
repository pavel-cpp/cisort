import os
import pathlib
import json

from config import FILE_EXTENTIONS

INCORRECT_JSON = 'File {path} is not json file'

def load_json(path: str):
  if pathlib.Path(path).suffix != '.json':
    raise ValueError(INCORRECT_JSON.format(path=path))
  with open(path) as file:
    return json.load(file)


def get_files_path(path: str):
    files_path = []
    for root, _, files in os.walk(path):
        for file in files:
            filepath = f'{root}/{file}'
            if pathlib.Path(filepath).suffix in FILE_EXTENTIONS:
                files_path.append(filepath)
    return files_path
