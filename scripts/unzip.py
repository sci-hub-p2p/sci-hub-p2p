import sys
import zipfile
from os import path
from urllib.request import urlopen


def fetch_torrent():
    zip_path = sys.argv[1]
    output = sys.argv[2]

    with zipfile.ZipFile(zip_path) as zip_ref:
        zip_ref.extractall(output)


if __name__ == "__main__":
    fetch_torrent()
