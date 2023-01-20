TESTDIR="testData/4-node-test1600"

# Start nodes
./node/node -conf=$TESTDIR/nodes-0.txt -cpu=1 --loglevel=5 &
./node/node -conf=$TESTDIR/nodes-1.txt -cpu=1 --loglevel=5 &
./node/node -conf=$TESTDIR/nodes-2.txt -cpu=1 --loglevel=5 &
./node/node -conf=$TESTDIR/nodes-3.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-4.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-5.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-6.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-7.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-8.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-9.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-10.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-11.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-12.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-13.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-14.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-15.txt -cpu=1 --loglevel=5 &

sleep 2

./client/client --conf=$TESTDIR/client.txt -metric=1 -batch=$1

killall ./node/node
