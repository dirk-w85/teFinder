#!/usr/bin/bash
archs=(amd64 arm64 ppc64le ppc64 s390x)
oss=(darwin linux)

for oss in ${oss[@]}
do
  for arch in ${archs[@]}
  do
      echo "Building for "${arch}
	  env GOOS=linux GOARCH=${arch} go build -o teFinder_${oss}_${arch}
  done
done