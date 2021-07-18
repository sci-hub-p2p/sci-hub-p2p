#!/usr/bin/env bash

#curl https://libgen.rs/scimag/repository_torrent/sm_00900000-00999999.torrent \
#  -o ./testdata/sm_00900000-00999999.torrent
#

#!/usr/bin/env bash

FILES="$HOME/repository_torrent/sm_*.torrent"
for f in $FILES; do
  echo "Processing $f file..."
  ./dist/sci-hub_windows_64.exe indexes gen -t "$f" -d '\\OMV\3t\sci-hub' -n 1
done
