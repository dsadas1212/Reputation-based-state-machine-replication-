.PHONY: all clean client node

all: client node

client: 
	make -C client

node:
	make -C node

clean:
	make -C client clean
	make -C node clean