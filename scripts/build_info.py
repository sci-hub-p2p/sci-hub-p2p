import re
import os
from datetime import datetime

ref = os.getenv("GITHUB_REF", "develop")
SHA = os.getenv("GITHUB_SHA", "00000000")[:8]
build_time = datetime.utcnow().replace(microsecond=0).isoformat()

if ref.startswith(
    (
        "refs/tags/",
        "refs/heads/",
    )
):
    ref = "".join(ref.split("/")[2:])
elif match := re.match("refs/pull/(.*)/merge", ref):
    ref = "pr-" + match.group(1)

print(
    f"""
REF="{ref}"
SHA="{SHA}"
TIME="{build_time}"
"""
)
