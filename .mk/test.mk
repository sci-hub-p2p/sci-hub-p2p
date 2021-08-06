testdata/sm_00900000-00999999.torrent: scripts/fetch_torrent.py
	python scripts/fetch_torrent.py

testdata/big_file.bin: ./scripts/gen_big_file.py
	python ./scripts/gen_big_file.py

testdata: testdata/sm_00900000-00999999.torrent testdata/big_file.bin $(GoSrc)

test: testdata
	go test -v ./...

coverage.out: testdata
	go test -covermode=atomic -coverprofile=coverage.out -count=1 ./...

coverage: coverage.out

clean::
	rm -rf testdata/sm_00900000-00999999.torrent testdata/big_file.bin coverage.out

.PHONY:: testdata test coverage clean
