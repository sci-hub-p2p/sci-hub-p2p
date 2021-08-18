"""replace makefile built-in `wildcard`.
Find file with save extensions recursively.

```bash
python scripts/wildcard.py go
```

```text
./cmd/cmd.go
./cmd/daemon/root.go
./cmd/flag/global.go
./cmd/indexes/index.go
```

"""

import os
import sys
import posixpath
from os import path


def should_skip(dir_name: str):
    s = path.normpath(dir_name).split(os.sep)
    for part in s:
        if part.startswith("."):
            return True
    return False


def main():
    ext = sys.argv[1]

    if not ext.startswith("."):
        ext = "." + ext

    for dir, _, files in os.walk("."):
        if should_skip(dir):
            continue
        for file in files:
            if file.endswith(ext):
                print(posixpath.join(*path.join(dir, file).split(os.sep)))


if __name__ == "__main__":
    main()
