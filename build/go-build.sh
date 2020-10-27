#!/usr/bin/env bash

target=$1
if [[ -z "$target" ]]; then
  echo "usage: $0 <directory> <output"
  exit 1
fi

output=$2
if [[ -z "$output" ]]; then
  echo "usage: $0 <directory> <output"
  exit 1
fi

platforms=("linux/amd64" "darwin/amd64")

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$output'-'$GOOS'-'$GOARCH

    echo "Building $output_name"
    env GOOS=$GOOS GOARCH=$GOARCH go build -o ./$target/$output_name ./$target
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done
