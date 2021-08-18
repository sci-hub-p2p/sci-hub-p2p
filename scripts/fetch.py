import os
import sys
from urllib.request import urlopen


def fetch_torrent():
    output = sys.argv[2]
    dst = os.path.join(os.path.dirname(__file__), "..", output)
    url = sys.argv[1]

    with urlopen(url) as response:
        content = response.read()

    os.makedirs(os.path.dirname(dst), exist_ok=True)
    with open(dst, "wb") as file:
        file.write(content)


if __name__ == "__main__":
    fetch_torrent()
