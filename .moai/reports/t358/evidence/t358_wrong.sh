go test -count=1 -json ./internal/version/... > /tmp/t358_e_stream2.json
rc=$?
echo "CENSUS RAN (rc=$rc)"
exit $rc
