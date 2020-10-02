.PHONY: all testfiles

all: bin
	@echo "Finished building tools"

testfiles: bin test3 test10 b100 b400 b800 b2000 b5000
	@echo "Built test files"

bin: genConfig.go options.go
	go build -o genConfig .

test3: bin
	mkdir -p ../testData/3-node-test/
	./genConfig -o ../testData/3-node-test/ -n3 -d50

test10: bin
	mkdir -p ../testData/10-node-test/
	./genConfig -o ../testData/10-node-test/ -n10 -d50

b100: bin
	mkdir -p ../testData/b100/
	./genConfig -o ../testData/b100/ -n3 -d50 -b100

b400: bin
	mkdir -p ../testData/b400/
	./genConfig -o ../testData/b400/ -n3 -d50 -b400

b800: bin
	mkdir -p ../testData/b800/
	./genConfig -o ../testData/b800/ -n3 -d50 -b800

b2000: bin
	mkdir -p ../testData/b2000/
	./genConfig -o ../testData/b2000/ -n3 -d50 -b2000

b5000: bin
	mkdir -p ../testData/b5000/
	./genConfig -o ../testData/b5000/ -n3 -d50 -b5000

clean: 
	rm -rf genConfig