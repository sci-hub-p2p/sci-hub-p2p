from os import path
from urllib.request import urlopen, Request

dst = path.join(path.dirname(__file__), "../testdata/sm_00900000-00999999.torrent")


def fetch_torrent():
    url = "https://libgen.rs/scimag/repository_torrent/sm_00900000-00999999.torrent"

    httprequest = Request(url)

    with urlopen(httprequest) as response:
        with open(dst, "wb") as file:
            file.write(response.read())


if __name__ == "__main__":
    fetch_torrent()
