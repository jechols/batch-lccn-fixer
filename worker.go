package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"sync"
)

// Workers just poll the job queue until it's done
type Worker struct {
	ID       int
	queue    chan *Job
	wg       *sync.WaitGroup
	badLCCN  []byte
	goodLCCN []byte
}

// Start listens for jobs until the work queue is closed
func (w *Worker) Start() {
	w.wg.Add(1)
	for j := range w.queue {
		log.Printf("DEBUG: worker %d Processing %s Job for %q", w.ID, j.Type, j.DestPath)
		switch j.Type {
		case XMLFix:
			w.FixXML(j)
		case PDFFix:
			w.FixPDF(j)
		case FileCopy:
			w.CopyFile(j)
		}
	}
	w.wg.Done()
}

// CopyFile just opens source and copies the contents to the destination path
func (w *Worker) CopyFile(j *Job) {
	var err = copyfile(j.SourcePath, j.DestPath)
	if err != nil {
		log.Printf("ERROR: %s", err)
	}
}

// FixXML reads the entire source file into memory, replaces all occurrences of
// the bad LCCN with the good LCCN, and writes the contents to the destination
func (w *Worker) FixXML(j *Job) {
	var b, err = ioutil.ReadFile(j.SourcePath)
	if err != nil {
		log.Printf("ERROR: unable to read %q: %s", j.SourcePath, err)
		return
	}

	var newBytes = bytes.Replace(b, w.badLCCN, w.goodLCCN, -1)
	err = ioutil.WriteFile(j.DestPath, newBytes, 0644)
	if err != nil {
		log.Printf("ERROR: unable to write %q: %s", j.DestPath, err)
	}
}
