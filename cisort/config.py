import argparse
import logging
from logging.handlers import RotatingFileHandler

from constants import FILE_CASE, LOG_DIR, LOG_FILE, PRETTY_CASE

LOG_FORMAT = '"[%(levelname)s] %(asctime)s - %(message)s"'
DT_FORMAT = '%d.%m.%Y %H:%M:%S'


def configure_argument_parser():
        parser = argparse.ArgumentParser(description='Includes sorter C++')
        parser.add_argument(
            'path',
            help='Directory or filename to sort'
        )
        parser.add_argument(
            '-v',
            '--verbose',
            action='store_true',
            help='Show more information about sorting process'
        )
        parser.add_argument(
            '-r',
            '--recursive',
            action='store_true',
            help='Switch mode to recursive searching files to sort'
        )
        parser.add_argument(
            '-c',
            '--comments',
            action='store_true',
            help='Add name of include group'
        )
        return parser

def configure_logging():
    LOG_DIR.mkdir(exist_ok=True)
    
    rotating_handler = RotatingFileHandler(
        LOG_FILE, maxBytes=10 ** 6, backupCount=5
    )
    logging.basicConfig(
        datefmt=DT_FORMAT,
        format=LOG_FORMAT,
        level=logging.INFO,
        handlers=(rotating_handler, logging.StreamHandler())
    )
