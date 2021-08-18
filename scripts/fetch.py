import sys
from os import path
from urllib.request import urlopen


def fetch_torrent():
    output = sys.argv[2]
    dst = path.join(path.dirname(__file__), "..", output)
    url = sys.argv[1]

    with open(dst, "wb") as file:
        with urlopen(url) as response:
            file.write(response.read())


if __name__ == "__main__":
    fetch_torrent()
