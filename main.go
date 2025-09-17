package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/iceber/iouring-go"
)

const entries uint = 64
const blockSize uint64 = 1024 * 1024 // 1MB

func main() {
	// check args
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s file1 file2\n", os.Args[0])
		return
	}

	// new IOURing
	// +1 for the last fsync request
	iour, err := iouring.New(entries + 1)
	if err != nil {
		panic(fmt.Sprintf("new IOURing error: %v", err))
	}
	defer iour.Close()

	fmt.Printf("open src file: %s\n", os.Args[1])
	var srcBytes []byte
	var size uint64
	var sizeMB float64
	{
		srcBytes, err = os.ReadFile(os.Args[1])
		if err != nil {
			fmt.Printf("Open src file failed: %v\n", err)
			return
		}

		stat, err := os.Stat(os.Args[1])
		if err != nil {
			panic(err)
		}
		size = uint64(stat.Size())
		sizeMB = float64(size) / 1024 / 1024

		if (size % blockSize) != 0 {
			panic("size is not a multiple of blockSize")
		}
	}

	fmt.Printf("create dest file: %s\n", os.Args[2])
	dest, err := os.OpenFile(os.Args[2], os.O_RDWR|os.O_CREATE|os.O_TRUNC|syscall.O_DIRECT, 0666)
	if err != nil {
		fmt.Printf("create dest file failed: %v\n", err)
		return
	}
	defer dest.Close()

	fd := int(dest.Fd())

	// register files
	if err := iour.RegisterFile(dest); err != nil {
		panic(err)
	}

	numBlocks := int(size / blockSize)
	numBatches := int(numBlocks / int(entries))
	fmt.Printf("numBatches: %d\n", numBatches)

	ch := make(chan iouring.Result, entries)
	fmt.Printf("write requests\n")
	var wg sync.WaitGroup

	{
		wg.Add(numBlocks)

		// add numBatches for the last fsync request
		wg.Add(numBatches)

		go func() {
			defer close(ch)
			for range numBlocks + numBatches {
				result := <-ch
				if err := result.Err(); err != nil {
					panic(err)
				}
				// fmt.Printf("batch %d done\n", batch)
				wg.Done()
			}
		}()
	}

	start := time.Now()

	{
		fmt.Printf("submit requests\n")
		for batch := range numBatches {
			// +1 for the last fsync request
			prepRequests := make([]iouring.PrepRequest, entries+1)
			batchOffset := uint64(batch) * blockSize * uint64(entries)

			for entry := range entries {
				offset := batchOffset + uint64(entry)*blockSize
				readSize := min(size-offset, blockSize)

				b := srcBytes[offset : offset+readSize]
				prepRequests[entry] = iouring.Pwrite(fd, b, offset)
			}

			prepRequests[entries] = iouring.Fsync(fd)

			if _, err := iour.SubmitRequests(prepRequests, ch); err != nil {
				panic(err)
			}
			// fmt.Printf("batch %d submitted\n", batch)
		}
	}

	fmt.Printf("wait for requests to complete\n")
	wg.Wait()

	elapsed := time.Since(start)
	bytePerSecond := sizeMB / elapsed.Seconds()
	fmt.Printf("byte per second: %f MB/s\n", bytePerSecond)
	fmt.Printf("elapsed: %vs\n", elapsed.Seconds())
}
