package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// JobType is an enumeration of the different types of jobs we handle
type JobType int

const (
	Unknown JobType = iota
	XMLFix
	PDFFix
	FileCopy
)

func (jt JobType) String() string {
	switch jt {
	case XMLFix:
		return "XML fix"
	case PDFFix:
		return "PDF fix"
	case FileCopy:
		return "File copy"
	}
	return "Unknown"
}

// FileRequest holds the source and destination paths for a file which needs to
// be processed, and the type of processing it needs
type Job struct {
	SourcePath string
	DestPath   string
	Type       JobType
	Failures   int
}

// The WorkQueue holds the workers and allows adding jobs and stopping the job
// collection process
type WorkQueue struct {
	workers []*Worker
	queue   chan *Job
	wg      *sync.WaitGroup
}

// NewWorkQueue creates n workers and starts them listening for jobs
func NewWorkQueue(ctx *FixContext, n int) *WorkQueue {
	var q = &WorkQueue{
		workers: make([]*Worker, n),
		queue:   make(chan *Job, 100000),
		wg:      new(sync.WaitGroup),
	}

	for i := 0; i < n; i++ {
		q.workers[i] = &Worker{
			ID:       i,
			queue:    q.queue,
			wg:       q.wg,
			badLCCN:  []byte(ctx.BadLCCN),
			goodLCCN: []byte(ctx.GoodLCCN),
		}
		go q.workers[i].Start()
	}

	return q
}

func (q *WorkQueue) Add(sourcePath, destDir, baseName string) {
	// Create the destination directory if it doesn't exist
	var err = os.MkdirAll(destDir, 0755)
	if err != nil {
		log.Printf("ERROR: could not create %q: %s", destDir, err)
		return
	}

	var ext = strings.ToLower(filepath.Ext(baseName)[1:])
	log.Printf("INFO: analyzing %q (destination %q, type %s)", sourcePath, destDir, ext)
	var destFile = filepath.Join(destDir, baseName)
	var job = &Job{SourcePath: sourcePath, DestPath: destFile}

	if len(baseName) > 10 && ext == "xml" {
		job.Type = XMLFix
	} else if ext == "pdf" {
		job.Type = PDFFix
	} else {
		job.Type = FileCopy
	}

	q.queue <- job
}

// Wait blocks until the queue is empty and all workers have quit
func (q *WorkQueue) Wait() {
	for _, w := range q.workers {
		w.Done()
	}
	q.wg.Wait()
}
