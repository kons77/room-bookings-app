testall: 
	@echo running all tests 
	@go test -v ./... -coverprofile="coverage.out" || echo "test fails"
	@go tool cover -html="coverage.out"

runw:
	@echo build and run app on Windons 
	@go build -o .\dist\bookings.exe ./cmd/web 
	@set CACHE=false&& start /B .\dist\bookings.exe -cache=false -production=false
	@echo booking running 

stopw:
	@taskkill /IM bookings.exe /F
	@echo booking stopper

runl:
	@echo build and run app on Linux
	@go build -o bookings cmd/web/*.go
	@./bookings  -dbname=bookings -dbuser=postgres -cache=false -production=false

start: runw

stop: stopw