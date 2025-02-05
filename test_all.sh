go test -v ./... -json | tparse
go test -v ./... -coverprofile="coverage.out" 