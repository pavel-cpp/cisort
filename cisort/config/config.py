import argparse
import logging
from logging.handlers import RotatingFileHandler


LOG_FORMAT = '"[%(levelname)s] %(asctime)s - %(message)s"'
DT_FORMAT = '%d.%m.%Y %H:%M:%S'


def configure_argument_parser():
    parser = argparse.ArgumentParser(description='Includes sorter C++')
    parser.add_argument(
            'files',
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
    parser.add_argument(
            '-g',
            '--group',
            action='store_true',
            help='Group includes'
    )
    return parser

def configure_logging():
    logging.basicConfig(
            datefmt=DT_FORMAT,
            format=LOG_FORMAT,
            level=logging.INFO,
            handlers=(logging.StreamHandler(),)
    )
