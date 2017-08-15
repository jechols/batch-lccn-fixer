package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
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
}

// Workers just poll the job queue until it's done
type Worker struct {
	ID    int
	queue chan *Job
}

// The WorkQueue holds the workers and allows adding jobs and stopping the job
// collection process
type WorkQueue struct {
	workers  []*Worker
	queue    chan *Job
	badLCCN  string
	goodLCCN string
}

// NewWorkQueue creates n workers and starts them listening for jobs
func NewWorkQueue(ctx *FixContext, n int) *WorkQueue {
	var wq = &WorkQueue{
		workers:  make([]*Worker, n),
		queue:    make(chan *Job, 100000),
		badLCCN:  ctx.BadLCCN,
		goodLCCN: ctx.GoodLCCN,
	}

	for i := 0; i < n; i++ {
		wq.workers[i] = &Worker{ID: i, queue: wq.queue}
		go wq.workers[i].Start()
	}

	return wq
}

func (q *WorkQueue) Add(sourcePath, destDir, baseName string) {
	// Create the destination directory if it doesn't exist
	var err = os.MkdirAll(destDir, 0755)
	if err != nil {
		// TODO: Add a results or errors channel!
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

// Start listens for jobs until the work queue is closed
func (w *Worker) Start() {
	for j := range w.queue {
		log.Printf("DEBUG: worker %d Processing %s Job for %q", w.ID, j.Type, j.DestPath)
		switch j.Type {
		case XMLFix:
		case PDFFix:
		case FileCopy:
			w.CopyFile()
		}
	}
}

func (w *Worker) CopyFile() {
}
