package main

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

const blockSize uint64 = 1024 * 1024 // 1MB

func main() {
	// check args
	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s file1 file2\n", os.Args[0])
		return
	}

	fmt.Printf("open src file: %s\n", os.Args[1])
	var srcBytes []byte
	var size uint64
	var sizeMB float64
	{
		var err error
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

	start := time.Now()

	fmt.Println("write data to dest file")
	{
		times := int(size/blockSize + 1)
		for i := range times {
			offset := uint64(i) * blockSize
			readSize := min(size-offset, blockSize)

			b := srcBytes[offset : offset+readSize]
			_, err := dest.Write(b)
			if err != nil {
				fmt.Printf("write failed: %v\n", err)
				return
			}
		}
	}

	elapsed := time.Since(start)
	bytePerSecond := sizeMB / elapsed.Seconds()
	fmt.Printf("byte per second: %f MB/s\n", bytePerSecond)
	fmt.Printf("elapsed: %vs\n", elapsed.Seconds())
}
