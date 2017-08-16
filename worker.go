package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

// Workers just poll the job queue until it's done
type Worker struct {
	sync.Mutex
	ID       int
	queue    chan *Job
	wg       *sync.WaitGroup
	badLCCN  []byte
	goodLCCN []byte
	done     bool
}

// Start listens for jobs until the work queue is closed
func (w *Worker) Start() {
	w.wg.Add(1)
	for {
		select {
		case j := <-w.queue:
			w.process(j)
		default:
			w.Lock()
			var isDone = w.done
			w.Unlock()
			if isDone {
				w.wg.Done()
				log.Printf("INFO: worker %d exiting", w.ID)
				return
			}
			time.Sleep(time.Millisecond * 50)
		}
	}
}

func (w *Worker) Done() {
	log.Printf("DEBUG: worker %d draining pool and exiting")
	w.Lock()
	w.done = true
	w.Unlock()
}

func (w *Worker) process(j *Job) {
	if j.Failures > 0 {
		log.Printf("DEBUG: worker %d Processing %s Job for %q (retry #%d)", w.ID, j.Type, j.DestPath, j.Failures)
	} else {
		log.Printf("DEBUG: worker %d Processing %s Job for %q", w.ID, j.Type, j.DestPath)
	}
	switch j.Type {
	case XMLFix:
		w.FixXML(j)
	case PDFFix:
		w.FixPDF(j)
	case FileCopy:
		w.CopyFile(j)
	}
}

// retry will put the job back into the main queue unless it has already failed
// too many times
func (w *Worker) retry(j *Job, reasonFormat string, reasonArgs ...interface{}) {
	var reason = fmt.Sprintf(reasonFormat, reasonArgs...)
	j.Failures++
	if j.Failures >= 5 {
		log.Printf("ERROR: %s", reason)
		return
	}

	log.Printf("WARN: %s; trying again (retry #%d)", reason, j.Failures)
	w.queue <- j
}

// CopyFile just opens source and copies the contents to the destination path
func (w *Worker) CopyFile(j *Job) {
	var err = copyfile(j.SourcePath, j.DestPath)
	if err != nil {
		w.retry(j, err.Error())
	}
}

// FixXML reads the entire source file into memory, replaces all occurrences of
// the bad LCCN with the good LCCN, and writes the contents to the destination
func (w *Worker) FixXML(j *Job) {
	var b, err = ioutil.ReadFile(j.SourcePath)
	if err != nil {
		w.retry(j, "unable to read %q: %s", j.SourcePath, err)
		return
	}

	var newBytes = bytes.Replace(b, w.badLCCN, w.goodLCCN, -1)
	err = ioutil.WriteFile(j.DestPath, newBytes, 0644)
	if err != nil {
		w.retry(j, "unable to write %q: %s", j.DestPath, err)
	}
}
