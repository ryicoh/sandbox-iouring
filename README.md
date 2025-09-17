# Disk I/O Benchmark with io_uring

A Go-based disk I/O benchmark tool that uses io_uring for high-performance asynchronous file operations. This tool measures write performance by copying a 32GB file using io_uring's efficient I/O submission and completion mechanism.

## Usage

```bash
# Generate a 32GB test file
$ make generate-32gb

# Run the benchmark
$ make run
go run ./main.go ./input_32G.bin output_32G.bin
open src file: ./input_32G.bin
create dest file: output_32G.bin
numBatches: 512
write requests
submit requests
wait for requests to complete
byte per second: 4166.654410 MB/s
elapsed: 7.864343133s

# Check the output file
$ make checksum
```

The benchmark will output the write speed in MB/s and the total elapsed time.