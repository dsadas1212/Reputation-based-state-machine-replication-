.PHONY: all testfiles

all: bin
	@echo "Finished building tools"

testfiles: bin test3 test10 b100 b400 b800 b2000 b5000 test41 test42 test43 test44 test81 test82 test83 test84 test161 test162 test163 test164
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
	
test42: bin
	mkdir -p ../testData/4-node-test400/
	./genConfig -o ../testData/4-node-test400/ -n4 -f0 -d50 -b400

test41: bin
	mkdir -p ../testData/4-node-test100/
	./genConfig -o ../testData/4-node-test100/ -n4 -f0 -d50 -b100	

test43: bin
	mkdir -p ../testData/4-node-test800/
	./genConfig -o ../testData/4-node-test800/ -n4 -f0 -d50 -b800

test44: bin
	mkdir -p ../testData/4-node-test2000/
	./genConfig -o ../testData/4-node-test2000/ -n4 -f0 -d50 -b2000

test81: bin
	mkdir -p ../testData/8-node-test100/
	./genConfig -o ../testData/8-node-test100/ -n8 -f0 -d50 -b100

test82: bin
	mkdir -p ../testData/8-node-test400/
	./genConfig -o ../testData/8-node-test400/ -n8 -f0 -d50 -b400

test83: bin
	mkdir -p ../testData/8-node-test800/
	./genConfig -o ../testData/8-node-test800/ -n8 -f0 -d50 -b800
	
test84: bin
	mkdir -p ../testData/8-node-test2000/
	./genConfig -o ../testData/8-node-test2000/ -n8 -f0 -d50 -b2000

test161: bin
	mkdir -p ../testData/16-node-test100/
	./genConfig -o ../testData/16-node-test100/ -n16 -f0 -d50 -b100

test162: bin
	mkdir -p ../testData/16-node-test400/
	./genConfig -o ../testData/16-node-test400/ -n16 -f0 -d50 -b400

test163: bin
	mkdir -p ../testData/16-node-test800/
	./genConfig -o ../testData/16-node-test800/ -n16 -f0 -d50 -b800

test164: bin
	mkdir -p ../testData/16-node-test2000/
	./genConfig -o ../testData/16-node-test2000/ -n16 -f0 -d50 -b2000

test321: bin
	mkdir -p ../testData/32-node-test100/
	./genConfig -o ../testData/32-node-test100/ -n32 -f0 -d50 -b100

test322: bin
	mkdir -p ../testData/32-node-test400/
	./genConfig -o ../testData/32-node-test400/ -n32 -f0 -d50 -b400

test323: bin
	mkdir -p ../testData/32-node-test800/
	./genConfig -o ../testData/32-node-test800/ -n32 -f0 -d50 -b800

test324: bin
	mkdir -p ../testData/32-node-test2000/
	./genConfig -o ../testData/32-node-test2000/ -n32 -f0 -d50 -b2000

clean: 
	rm -rf genConfig