"""generate a test binary file at `./testdata/big_file.bin`"""
import sys
import random
import hashlib

assert sys.version_info >= (3, 6, 0), "use python >= 3.6"

# lets make it 60MB currently
# should be a very big PDF
filesize = 60 * 1024 * 1024
filepath = "./testdata/big_file.bin"
expected_sha256 = "a5e4d5b214589333198cd124aef844624f3c3ec4d29f0b1646cda2ff8c08d530"
generator = random.Random("seed")

basic = bytes(generator.getrandbits(8) for _ in range(filesize))
with open(filepath, "wb") as f:
    f.write(basic)
    f.truncate()

with open(filepath, "rb") as f:
    content = f.read()

content_sha256 = hashlib.sha256(content).hexdigest()

assert content_sha256 == expected_sha256, (
    "generated data hash different sha256, expect: " + content_sha256
)
