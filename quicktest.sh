TESTDIR="testData/3-node-test400"

# Start nodes
./node/node -conf=$TESTDIR/nodes-0.txt -cpu=1 --loglevel=5 &
./node/node -conf=$TESTDIR/nodes-1.txt -cpu=1 --loglevel=5 &
./node/node -conf=$TESTDIR/nodes-2.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-3.txt -cpu=1 --loglevel=5 &
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
# ./node/node -conf=$TESTDIR/nodes-16.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-17.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-18.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-19.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-20.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-21.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-22.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-23.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-24.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-25.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-26.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-27.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-28.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-29.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-30.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-31.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-32.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-33.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-34.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-35.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-36.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-37.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-38.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-39.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-40.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-41.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-42.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-43.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-44.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-45.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-46.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-47.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-48.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-49.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-50.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-51.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-52.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-53.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-54.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-55.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-56.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-57.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-58.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-59.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-60.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-61.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-62.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-63.txt -cpu=1 --loglevel=5 &
# ./node/node -conf=$TESTDIR/nodes-64.txt -cpu=1 --loglevel=5 &


sleep 2

./client/client --conf=$TESTDIR/client.txt -metric=1 -batch=$1

killall ./node/node
