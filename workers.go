package main

import (
	"io"
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
}

// Workers just poll the job queue until it's done
type Worker struct {
	ID    int
	queue chan *Job
	wg    *sync.WaitGroup
}

// The WorkQueue holds the workers and allows adding jobs and stopping the job
// collection process
type WorkQueue struct {
	workers  []*Worker
	queue    chan *Job
	badLCCN  string
	goodLCCN string
	wg *sync.WaitGroup
}

// NewWorkQueue creates n workers and starts them listening for jobs
func NewWorkQueue(ctx *FixContext, n int) *WorkQueue {
	var q = &WorkQueue{
		workers:  make([]*Worker, n),
		queue:    make(chan *Job, 100000),
		badLCCN:  ctx.BadLCCN,
		goodLCCN: ctx.GoodLCCN,
		wg:       new(sync.WaitGroup),
	}

	for i := 0; i < n; i++ {
		q.workers[i] = &Worker{ID: i, queue: q.queue, wg: q.wg}
		go q.workers[i].Start()
	}

	return q
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

// Wait blocks until the queue is empty; calls to Add() will fail at this point
func (q *WorkQueue) Wait() {
	close(q.queue)
	q.wg.Wait()
}

// Start listens for jobs until the work queue is closed
func (w *Worker) Start() {
	w.wg.Add(1)
	for j := range w.queue {
		log.Printf("DEBUG: worker %d Processing %s Job for %q", w.ID, j.Type, j.DestPath)
		switch j.Type {
		case XMLFix:
		case PDFFix:
		case FileCopy:
			w.CopyFile(j)
		}
	}
	w.wg.Done()
}

// CopyFile just opens source and copies the contents to the destination path.
// There is an inordinate amount of error handling because if something goes
// wrong we *really* need to know exactly what it was.
func (w *Worker) CopyFile(j *Job) {
	var in, out *os.File
	var err error

	in, err = os.Open(j.SourcePath)
	if err != nil {
		log.Printf("ERROR: unable to read %q: %s", j.SourcePath, err)
		return
	}
	defer in.Close()

	out, err = os.Create(j.DestPath)
	if err != nil {
		log.Printf("ERROR: unable to create %q: %s", j.DestPath, err)
		return
	}

	_, err = io.Copy(out, in)
	if err != nil {
		log.Printf("ERROR: unable to write to %q: %s", j.DestPath, err)
		return
	}

	err = out.Sync()
	if err != nil {
		log.Printf("ERROR: unable to sync %q: %s", j.DestPath, err)
		return
	}

	err = out.Close()
	if err != nil {
		log.Printf("ERROR: unable to close %q: %s", j.DestPath, err)
		return
	}
}
