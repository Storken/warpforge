mkdir -p $HOME/.warpforge/warehouse

# add alpine image for testing
./plugins/rio mirror --source=file://plugins/alpine.tgz --target=ca+file://$HOME/.warpforge/warehouse tar:47Yg1Sdq21rPyDw9X9sCmRubQUADhFKe9G7qZCJRe61RhWPCxcQysCFzyCHffBKRjB

# add empty tarball for testing
./plugins/rio mirror --source=file://plugins/test.tgz --target=ca+file://$HOME/.warpforge/warehouse tar:7omHHaRUV3TcPYLk7VWTQgFSAWJa3HTRVwiZwESBy65w8rbrtVqdtZPg2nL1zXWPmR


mkdir -p $HOME/.warpforge/bin
cp plugins/rio $HOME/.warpforge/bin/
cp plugins/runc $HOME/.warpforge/bin/
