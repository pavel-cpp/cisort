import json
import os
from pathlib import Path
from typing import Iterable

from config import FILE_EXTENTIONS

INCORRECT_JSON = 'File {path} is not json file'


def load_json(path: str):
    if Path(path).suffix != '.json':
        raise ValueError(INCORRECT_JSON.format(path=path))
    with open(path) as file:
        return json.load(file)

def find_files(paths: Iterable[str]):
    for path in paths:
        if os.path.isdir(path):
            for root, _, files in os.walk(path):
                path = Path(root)
                correct_files = []
                for file in files:
                    filepath = path / file
                    if Path(filepath).suffix in FILE_EXTENTIONS:
                        correct_files.append(filepath)
                yield correct_files
        else:
            yield path
