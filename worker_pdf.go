package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
)

var sourceFileReplacer = regexp.MustCompile(`"SourceFile":\s*"[^"]+"`)

// FixPDF copies the file to the destination location and shells out to
// exiftool to read and replace the fields in the PDF in order to swap the bad
// LCCN for the good.
func (w *Worker) FixPDF(j *Job) {
	// We work on a copy of the file, not the original!
	var err = copyfile(j.SourcePath, j.DestPath)
	if err != nil {
		w.retry(j, "unable to copy PDF file %q: %s", j.SourcePath, err)
		return
	}

	// Gather intel on the file
	var cmd = exec.Command("exiftool", "-json", j.DestPath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		w.retry(j, "unable to get EXIF data for %q: %s", j.DestPath, err)
		return
	}

	// Fix it!  This may be a bit dangerous, but the odds of the bad LCCN
	// actually showing up in EXIF data and *not* needing to be changed seem
	// pretty unlikely.
	var fixed = bytes.Replace(out.Bytes(), w.badLCCN, w.goodLCCN, -1)

	// Forcibly set source file info; this just makes things a pain when the
	// filename has the old LCCN, and the old is a subset of the new
	var sourceLine = fmt.Sprintf(`"SourceFile": "%s"`, j.DestPath)
	fixed = sourceFileReplacer.ReplaceAllLiteral(fixed, []byte(sourceLine))

	// Create a temp file for storing the exif JSON
	var tmp *os.File
	tmp, err = ioutil.TempFile("", "")
	if err != nil {
		w.retry(j, "unable to create tempfile to store EXIF JSON for %q: %s", j.DestPath, err)
		return
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name())
	}()

	// Write the fixed data to a file; exiftool really wants its JSON from a
	// file, not stdin or flags
	var n int
	n, err = tmp.Write(fixed)
	if err != nil {
		w.retry(j, "unable to create tempfile to store EXIF JSON for %q: %s", j.DestPath, err)
		return
	}
	if n != len(fixed) {
		w.retry(j, "unable to create tempfile to store EXIF JSON for %q: data was only partially written", j.DestPath, err)
		return
	}

	// Import the fixed data
	cmd = exec.Command("exiftool", "-overwrite_original", "-json="+tmp.Name(), j.DestPath)
	err = cmd.Run()
	if err != nil {
		w.retry(j, "unable to write EXIF data for %q: %s", j.DestPath, err)
	}
}
