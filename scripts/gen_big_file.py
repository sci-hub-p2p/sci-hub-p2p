import sys
import random
import hashlib

assert sys.version_info >= (3, 6), "please use newer python version 3.6+"

generator = random.Random("seed")
expected_sha256 = "a5e4d5b214589333198cd124aef844624f3c3ec4d29f0b1646cda2ff8c08d530"
filepath = "./testdata/big_file.bin"

with open(filepath, "wb") as f:
    # lets make it 60MB currently
    # should be a very big PDF
    basic = bytes(generator.getrandbits(8) for _ in range(60 * 1024 * 1024))
    # write binary to bypass EOL difference
    f.write(basic)
    f.truncate()

with open(filepath, "rb") as f:
    content = f.read()

content_sha256 = hashlib.sha256(content).hexdigest()

assert content_sha256 == expected_sha256, (
    "generated data hash different sha256, expect: " + content_sha256
)
