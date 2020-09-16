.PHONY: all clean client node testdata proto

all: client node

client: 
	make -C client

node:
	make -C node

testdata:
	make -C tools
	make -C tools testdata

clean:
	make -C client clean
	make -C node clean
	
proto:
	make -C proto clean
	make -C proto