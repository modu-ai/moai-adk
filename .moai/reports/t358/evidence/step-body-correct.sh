rc=0
go test -json -coverprofile=coverage.out -covermode=atomic ./internal/hook/mx/complexity/... > test-stream.json || rc=$?
bash scripts/ci-census/test-census.sh test-stream.json
exit $rc
