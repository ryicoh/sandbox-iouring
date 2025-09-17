default:

run:
	go run ./main.go ./input_32G.bin output_32G.bin

generate-32gb:
	fio --name=generate-32gb --filename=input_32G.bin --size=32G --rw=write --bs=1M

checksum: checksum-32gb

checksum-32gb:
	$(MAKE) _checksum-32gb -j 2
	diff input_32G.bin.sha1 output_32G.bin.sha1

_checksum-32gb: checksum-input-32gb checksum-output-32gb

checksum-input-32gb:
	cat -n input_32G.bin | openssl dgst -sha1 | awk '{print $$2}' > input_32G.bin.sha1

checksum-output-32gb:
	cat -n output_32G.bin | openssl dgst -sha1 | awk '{print $$2}' > output_32G.bin.sha1