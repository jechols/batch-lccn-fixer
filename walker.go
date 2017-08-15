package main

import (
	"os"
	"path/filepath"
	"strings"
)

const Separator = string(os.PathSeparator)

// Walker is not a Texas ranger, just a structure for walking files with the
// context we need to know how to start jobs and translate a source file to
// where we'll want it to end up
type Walker struct {
	ctx *FixContext
	queue   *WorkQueue
}

func NewWalker(ctx *FixContext, queue *WorkQueue) *Walker {
	return &Walker{ctx, queue}
}

func (w *Walker) Walk() error {
	return filepath.Walk(w.ctx.SourceDir, w.walkFunc)
}

func (w *Walker) walkFunc(path string, info os.FileInfo, err error) error {
	// We don't do anything with directories
	if info.IsDir() {
		return nil
	}

	// Gather info
	var parts = strings.Split(path, Separator)
	var baseName = parts[len(parts)-1]
	var localDir = strings.Replace(strings.Replace(path, w.ctx.SourceDir, "", 1), baseName, "", 1)

	// Fix LCCN in local path pieces for determining the destination path
	var localDirParts = strings.Split(localDir, Separator)
	for i, p := range localDirParts {
		if p == w.ctx.BadLCCN {
			localDirParts[i] = w.ctx.GoodLCCN
		}
	}
	localDir = filepath.Join(localDirParts...)
	var destPath = filepath.Join(w.ctx.DestDir, localDir)

	// Queue it up and let the workers handle the rest
	w.queue.Add(path, destPath, baseName)

	return nil
}
