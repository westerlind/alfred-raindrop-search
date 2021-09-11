#!/bin/sh

## You have to also install Go first for this to work, for example with
## the following command, which in turn requires you to install
## Homebrew first, which you can do with the instructions at https://brew.sh
#  brew install go
## You then need to install Deanishe's great Go Alfred library "AwGo" with this command
#  go get -u github.com/deanishe/awgo
## Then you can go on and actually run this script to compile the code

# Build raindrop_alfred as Intel and ARM binaries, and combine into a universal binary
if [ -e raindrop_alfred ]
then
  rm raindrop_alfred
fi
GOOS=darwin GOARCH=amd64 go build -o raindrop_alfred_amd64 raindrop_main.go raindrop_common.go client_code.go raindrop_authserver.go raindrop_search.go raindrop_add.go
GOOS=darwin GOARCH=arm64 go build -o raindrop_alfred_arm64 raindrop_main.go raindrop_common.go client_code.go raindrop_authserver.go raindrop_search.go raindrop_add.go
lipo -create -output raindrop_alfred raindrop_alfred_amd64 raindrop_alfred_arm64
rm raindrop_alfred_amd64
rm raindrop_alfred_arm64