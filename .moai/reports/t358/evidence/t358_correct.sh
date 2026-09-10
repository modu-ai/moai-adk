rc=0
go test -count=1 -json ./internal/version/... > /tmp/t358_e_stream.json || rc=$?
echo "CENSUS RAN (rc=$rc)"
exit $rc
