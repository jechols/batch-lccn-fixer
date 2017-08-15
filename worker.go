package main

import (
	"log"
	"io"
	"os"
	"sync"
)

// Workers just poll the job queue until it's done
type Worker struct {
	ID    int
	queue chan *Job
	wg    *sync.WaitGroup
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
