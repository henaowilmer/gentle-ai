//go:build darwin

package reviewtransaction

import "os"

func publishNoReplace(source, destination string) error {
	return os.Link(source, destination)
}

func replaceFileAtomic(source, destination string) error {
	return os.Rename(source, destination)
}
