// This file is part of sophon.
// Copyright alibaba-inc.com

package main

import "fmt"

func VersionMajor() int {
	return 0
}

func VersionMinor() int {
	return 0
}

func VersionRevision() int {
	return 1
}

func Version() string {
	return fmt.Sprintf("%v.%v.%v", VersionMajor(), VersionMinor(), VersionRevision())
}

func Signature() string {
	return "Talks"
}
