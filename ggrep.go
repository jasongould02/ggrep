package main

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"bytes"
	"strings"
	"sync"
	"path/filepath"
)


var wg sync.WaitGroup; // waitgroup for go routines

const BUFFER_SIZE = 4096;

// Option Flags
var ignoreCase bool;
var recursive bool; // is recursive search
var fileList = make(map[string]bool); // list of paths to each file, prevents duplicates

var messagePipe = make(chan string, 200);
func messagePipeline() {
	for {
		fmt.Println(<-messagePipe);
		wg.Done();
	}
}

// pm is for sending 'msg' which is the file path and 'str' which is the line containing the pattern
// to the messagePipe channel for printing by a seperate go routine
func pm(msg string, str []byte) {
	wg.Add(1); // in cases where all files are finished being searched before message pipe can finish printing all matches
	messagePipe <- msg + ":" + string(str); // append the match found output message to messagePipe channel
}

func main() {
	args := os.Args[1:];
	hasOptions := false;
	lastOptionIndex := 0;
	for i := 0; i < len(args); i++ {
		if args[i] == "." {
			args[i] = "./";
		}
		if strings.HasPrefix(args[i], "-") {
			hasOptions = true;
			switch args[i] {
				case "-ig": // ignore cases when searching
					lastOptionIndex = i;
					ignoreCase = true;
				case "-r": // search for pattern(s) recursively
					lastOptionIndex = i;
					recursive = true;
				default: // not a recognized argument option
			}
		}
	}
	pattern := args[(lastOptionIndex)];
	filenames := args[(lastOptionIndex+1):];
	if hasOptions {
		pattern = args[(lastOptionIndex+1)]; // the pattern appears directly after the options
		filenames = args[(lastOptionIndex+2):]; // specified files appear after the pattern
	}

	for _, f := range filenames {
		readDir(f, recursive); // files and folders from cmdline arguments, are then checked and added to fileList map
	}

	// Start searching for patterns in files
	fmt.Println("Number of files to search:", len(fileList));
	go messagePipeline();
	for key, _ := range fileList {
		if key == "ggrep" {
			fmt.Println("Skipping the ggrep executable.");
			continue;
		}
		wg.Add(1);
		go func(fns string, p string) { // Each file will be opened, searched and closed on its own go routine
			file, _ := os.Open(fns);
			searchFile(fns, []byte(p), bufio.NewReaderSize(file, BUFFER_SIZE));
			file.Close();
			wg.Done();
		}(key, pattern);
	}
	wg.Wait();
}

func readDir(path string, rec bool) { // adds all files to the fileList map
	files, _ := os.ReadDir(path);
	for _, f := range files {
		tpath := filepath.Clean(path + string(os.PathSeparator) + f.Name()); // grab file path
		if f.IsDir() {
			if rec == true { 
				fileList[tpath] = true; // add directory to listing
				readDir(tpath, rec); // search directory
				continue;
			}
		}
		fileList[tpath] = true;
	}
}

// Returns the number of times pattern appears inside the given file
// Also prints matches to the message pipe
func searchFile(filename string, patternByte []byte, reader io.Reader) int {
	count := 0;
	buffer := make([]byte, BUFFER_SIZE);
	totalBytes := 0;
	totalMatches := 0;
	//patternByte := []byte(pattern); // byte array of the pattern, 
	for {
		bufferSize, err := reader.Read(buffer);
		if err != nil && err != io.EOF { // end loop if error in reading 
			return 0;
		}
		var position int;
		for {
			i := bytes.IndexByte(buffer[position:], '\n');
			if i == -1 || bufferSize == position {
				break;
			}
			inLine := 0;
			if ignoreCase == true {
				inLine = bytes.Count(bytes.ToLower(buffer[position:position+i]), bytes.ToLower(patternByte));
			} else {
				inLine = bytes.Count(buffer[position:position+i], patternByte);
			}
			if inLine > 0 {
				pm(filename, buffer[position:position+i]); // send data to message pipe
			}
			totalMatches += inLine;
			position += i + 1; // move forward in file
			count += 1;
		}
		totalBytes += position;
		if err == io.EOF { // EOF after parsing buffer, so that last bit if file data is parsed
			break;
		}
	}
	return count;
}

