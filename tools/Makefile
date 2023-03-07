.PHONY: all testfiles

all: bin
	@echo "Finished building tools"

testfiles: bin test3 test31 test10 b100 b400 b800 b2000 b5000 test41 test42 test43 test44 test45 test46 test47 test9 test7 test33 test65
	@echo "Built test files"

bin: genConfig.go options.go
	go build -o genConfig .

test3: bin
	mkdir -p ../testData/3-node-test/
	./genConfig -o ../testData/3-node-test/ -n3 -d50

test31: bin
	mkdir -p ../testData/3-node-test1600/
	./genConfig -o ../testData/3-node-test1600/ -n3 -f1 -d50 -b1600

test32: bin
	mkdir -p ../testData/3-node-test800/
	./genConfig -o ../testData/3-node-test800/ -n3 -f1 -d50 -b800

test343: bin
	mkdir -p ../testData/3-node-test400/
	./genConfig -o ../testData/3-node-test400/ -n3 -f0 -d50 -b400

test344: bin
	mkdir -p ../testData/9-node-test400/
	./genConfig -o ../testData/9-node-test400/ -n9 -f0 -d50 -b400

test345: bin
	mkdir -p ../testData/17-node-test400/
	./genConfig -o ../testData/17-node-test400/ -n17 -f0 -d50 -b400

test346: bin
	mkdir -p ../testData/33-node-test400/
	./genConfig -o ../testData/33-node-test400/ -n33 -f0 -d50 -b400

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

test45: bin
	mkdir -p ../testData/4-node-test1600/
	./genConfig -o ../testData/4-node-test1600/ -n4 -f0 -d50 -b1600

test46: bin
	mkdir -p ../testData/4-node-test1800/
	./genConfig -o ../testData/4-node-test1800/ -n4 -f0 -d50 -b1800

test47: bin
	mkdir -p ../testData/4-node-test3000/
	./genConfig -o ../testData/4-node-test3000/ -n4 -f0 -d50 -b3000

#test performance in faluty nodes(4/9/17/33/65)

test9: bin
	mkdir -p ../testData/9-node-test1600/
	./genConfig -o ../testData/9-node-test1600/ -n9 -f4 -d50 -b800
test91: bin
	mkdir -p ../testData/9-node-test800/
	./genConfig -o ../testData/9-node-test800/ -n9 -f4 -d50 -b800

test17: bin
	mkdir -p ../testData/17-node-test800/
	./genConfig -o ../testData/17-node-test800/ -n17 -f8 -d50 -b800

test33: bin
	mkdir -p ../testData/33-node-test800/
	./genConfig -o ../testData/33-node-test800/ -n33 -f16 -d50 -b800

test65: bin
	mkdir -p ../testData/65-node-test800/
	./genConfig -o ../testData/65-node-test800/ -n65 -f32 -d50 -b800
clean: 
	rm -rf genConfig