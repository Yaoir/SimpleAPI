fib: fib.go api.go
	@go build -o fibserver fib.go api.go

vet:
	@go vet fib.go
	@go vet api.go

race:
	@go build -race -o fib fib.go

reset:
	@/bin/echo -ne "\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00\00" >/tmp/fibonacci_api_data

truncate:
	@echo -n >/tmp/fibonacci_api_data

# go to the start of the sequence.
rewind:
	@for n in `seq 95`; do fib p >/dev/null 2>&1; done

# go to the highest int64 in the sequence
wind ff:
	@for n in `seq 95`; do fib n >/dev/null 2>&1; done

hd:
	@hd /tmp/fibonacci_api_data
wc:
	@wc *.go

clean:
	@rm -f fib

back bak backup:
	@cp -a fib.go makefile push README.md throughput-test TODO .bak

