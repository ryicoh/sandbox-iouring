# Disk I/O Benchmark with io_uring

A Go-based disk I/O benchmark tool that uses io_uring for high-performance asynchronous file operations. This tool measures write performance by copying a 32GB file using io_uring's efficient I/O submission and completion mechanism.

## Usage

```bash
# Generate a 32GB test file
$ make generate-32gb

# Run the benchmark
$ go run ./main.go ./input_32G.bin output_32G.bin
open src file: ./input_32G.bin
create dest file: output_32G.bin
write requests
submit requests
byte per second: 2904.870204 MB/s
elapsed: 11.280366315s
```

The benchmark will output the write speed in MB/s and the total elapsed time.