import pathlib
import json

INCORRECT_JSON = 'File {path} is not json file'

def load_json(path: str):
  if pathlib.Path(path).suffix != '.json':
    raise ValueError(INCORRECT_JSON.format(path=path))
  with open(path) as file:
    return json.load(file)
