package main

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/iceber/iouring-go"
)

const entries uint = 128
const blockSize uint64 = 1024 * 1024 // 1MB

func main() {
	// check args
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s file1 file2\n", os.Args[0])
		return
	}

	// new IOURing
	iour, err := iouring.New(entries)
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
	}

	fmt.Printf("create dest file: %s\n", os.Args[2])
	dest, err := os.OpenFile(os.Args[2], os.O_RDWR|os.O_CREATE|os.O_TRUNC|syscall.O_DIRECT, 0666)
	if err != nil {
		fmt.Printf("create dest file failed: %v\n", err)
		return
	}
	defer dest.Close()

	// register files
	if err := iour.RegisterFile(dest); err != nil {
		panic(err)
	}

	start := time.Now()

	fmt.Printf("write requests\n")
	{
		times := int(size/blockSize + 1)
		var wg sync.WaitGroup
		wg.Add(times)

		ch := make(chan iouring.Result, entries)
		go func() {
			defer close(ch)
			for result := range ch {
				if err := result.Err(); err != nil {
					panic(err)
				}
				wg.Done()
			}
		}()

		fmt.Printf("submit requests\n")
		for i := range times {
			offset := uint64(i) * blockSize
			readSize := min(size-offset, blockSize)

			b := srcBytes[offset : offset+readSize]
			prepRequest := iouring.Pwrite(int(dest.Fd()), b, offset)
			if _, err := iour.SubmitRequest(prepRequest, ch); err != nil {
				panic(err)
			}
		}

		wg.Wait()
	}

	elapsed := time.Since(start)
	bytePerSecond := sizeMB / elapsed.Seconds()
	fmt.Printf("byte per second: %f MB/s\n", bytePerSecond)
	fmt.Printf("elapsed: %vs\n", elapsed.Seconds())
}
