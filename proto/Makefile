# PROTO_PATH=/home/haxolotl/go/src/github.com/adithyabhatkajake/libsynchs/proto
# GO_OUT_DIR=/home/haxolotl/go/src/
THIS_DIR=$(dir $(abspath $(firstword $(MAKEFILE_LIST))))
PROTO_PATH=$(THIS_DIR)
GO_OUT_DIR=$(THIS_DIR)../
# Reference on how to use $^, $? and other automatic variables
# https://www.gnu.org/software/make/manual/html_node/Automatic-Variables.html

# Build all the protobuf files
.PHONY: all clean

all: config/*.proto msg/*.proto ./*.proto chain/*.proto
	@echo "Using Proto Path: ${PROTO_PATH}"
	@echo "Using Go Out Directory: ${GO_OUT_DIR}"
	# Build only changed protobuf definitions
	# Compiling Protobuf tips
	# https://jbrandhorst.com/post/go-protobuf-tips/
	protoc $? -I${PROTO_PATH} --go_out=:${GO_OUT_DIR}

clean:
	@find .. -name "*.pb.go" -delete
	@echo "Cleaned all protobuf builds"