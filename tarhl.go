package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"flag"
	"path"
	"io/fs"
)

type WriteQueueEntry struct {
	Name string
	Pair string
	Content []byte
	Info fs.FileInfo
}

func isSameFile(existing fs.FileInfo, new fs.FileInfo) (bool) {
    if existing.Size() != new.Size() {
		return false
	}
	if existing.Mode() != new.Mode() {
		return false
	}
	if !existing.ModTime().Equal(new.ModTime()) {
		return false
	}
	return true
}

func fileWriter(queue chan *WriteQueueEntry, done chan bool) {
	written := 0
	linked := 0
	for next := range queue {
		if next.Pair != "" {
			pairStat, err := os.Lstat(next.Pair)
			if err != nil {
				log.Fatal(err)
			}
			if isSameFile(pairStat, next.Info) {
				// same file, prefer to hard link
				if err = os.Link(next.Pair, next.Name); err != nil {
					log.Fatal(err)
				}
				linked++
				continue
			}
			
		}
		err := os.WriteFile(next.Name, next.Content, next.Info.Mode().Perm())
		written++
		if err != nil {
			log.Fatal(err)
		}
		modTime := next.Info.ModTime()
		if err = os.Chtimes(next.Name, modTime, modTime); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("Written: %d files, linked: %d files\n", written, linked)
	done <- true
}

func main() {
	filePtr := flag.String("file", "", "Input tar file (required)")
	outputDirPtr := flag.String("dest", "", "Destination folder (required)")
	baseDirPtr := flag.String("base", "", "Base directory to hardlink to (optional)")

	flag.Parse()

	if *filePtr == "" || *outputDirPtr == "" {
		fmt.Printf("Input file and destination must be specified.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Open and iterate through the files in the archive.
	var tr *tar.Reader

	if *filePtr == "-" {
		tr = tar.NewReader(os.Stdin)
	} else {
		file, err := os.Open(*filePtr)
		if err != nil {
			log.Fatal(err)
		}
		tr = tar.NewReader(file)
	}

	err := os.MkdirAll(*outputDirPtr, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Setup file writer
	queue := make(chan *WriteQueueEntry, 100)
	writerDone := make(chan bool)

	go fileWriter(queue, writerDone)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		destName := path.Join(*outputDirPtr, hdr.Name)
		pairFileName := ""
		if *baseDirPtr != "" {
			pairFileName = path.Join(*baseDirPtr, hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeReg:
			fileInfo := hdr.FileInfo()
	
			data, err := io.ReadAll(tr)
			if err != nil {
				log.Fatal(err)
			}
		
			queueEntry := &WriteQueueEntry{Name: destName, Pair: pairFileName, Content: data, Info: fileInfo}
			queue <- queueEntry 
		case tar.TypeDir:
			fileInfo := hdr.FileInfo()
			err = os.MkdirAll(destName, fileInfo.Mode().Perm())
			if err != nil {
				log.Fatal(err)
			}
		case tar.TypeSymlink:
			err = os.Symlink(hdr.Linkname, destName)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	close(queue)

	<-writerDone
}
