package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func usageError(msg string, args ...interface{}) {
	var fmsg = fmt.Sprintf(msg, args...)
	fmt.Printf("\033[31;1mERROR: %s\033[0m\n", fmsg)
	fmt.Printf(`
Usage: %s <source directory> <destination directory> <bad LCCN> <good LCCN>

Finds files in a given source directory that need an LCCN fix.  This includes
the XML files as well as PDF metadata.  After fixes are applied, files are
copied to the destination directory.

The source directory should be either the pristine dark archive files or else
a copy of those files (TIFFs aren't necessary, however).  The destination
should be where the batch should live when it's ready for ingest.
`, os.Args[0])
	os.Exit(1)
}

// getArgs does some sanity-checking and sets the source/dest args
func getArgs() *FixContext {
	if len(os.Args) < 5 {
		usageError("Missing one or more arguments")
	}
	if len(os.Args) > 5 {
		usageError("Too many arguments supplied")
	}

	var fc = &FixContext{
		SourceDir: os.Args[1],
		DestDir:   os.Args[2],
		BadLCCN:   os.Args[3],
		GoodLCCN:  os.Args[4],
	}
	var err error
	fc.SourceDir, err = filepath.Abs(fc.SourceDir)
	if err != nil {
		usageError("Source (%s) is invalid: %s", fc.SourceDir, err)
	}
	fc.DestDir, err = filepath.Abs(fc.DestDir)
	if err != nil {
		usageError("Source (%s) is invalid: %s", fc.DestDir, err)
	}

	var info os.FileInfo
	info, err = os.Stat(fc.SourceDir)
	if err != nil {
		usageError("Source (%s) is invalid: %s", fc.SourceDir, err)
	}
	if !info.IsDir() {
		usageError("Source (%s) is invalid: not a directory", fc.SourceDir)
	}

	_, err = os.Stat(fc.DestDir)
	if err == nil || !os.IsNotExist(err) {
		usageError("Destination (%s) already exists", fc.DestDir)
	}

	return fc
}

func main() {
	var fixContext = getArgs()
	var queue = NewWorkQueue(fixContext, 2 * runtime.NumCPU())
	var walker = NewWalker(fixContext, queue)
	var err = walker.Walk()
	if err != nil {
		fmt.Printf("Error trying to copy/fix the batch: %s\n", err)
		os.Exit(1)
	}

	// HACK: We need to be sure the queue has had time to start getting filled
	// up, especially if the disk crawling is particularly slow
	time.Sleep(time.Second * 5)
	queue.Wait()
}
