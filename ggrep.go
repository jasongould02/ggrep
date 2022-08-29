package main

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"bytes"
	"strings"
	"sync"
	//"time"
	"path/filepath"
)


var wg sync.WaitGroup; // waitgroup

const BUFFER_SIZE = 4096;

// Option Flags
var ignoreCase bool;
var recursive bool;
var fileList = make(map[string]bool);

var messagePipe = make(chan string, 200);
func messagePipeline() {
	for {
		fmt.Println(<-messagePipe);
	}
}

func pm(msg string) {
	messagePipe <- msg;
}

func main() {
	args := os.Args[1:];

	go messagePipeline();

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
				case "-mp": // parse for multiple patterns
					lastOptionIndex = i;
					fmt.Println("-mp is not implemented yet");
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
		pattern = args[(lastOptionIndex+1)];
		filenames = args[(lastOptionIndex+2):];
	}

	for _, f := range filenames {
		readDir(f, recursive);
	}

	// Start searching for patterns in files
	fmt.Println("Number of files to search:", len(fileList));
	for key, _ := range fileList {
		if key == "ggrep" {
			fmt.Println("Skipping the ggrep executable."); // Don't forget to remove this later.
			continue;
		}
		wg.Add(1);
		go func(fns string, p string) {
			file, _ := os.Open(fns);
			searchFile(fns, p, bufio.NewReaderSize(file, BUFFER_SIZE));
			file.Close();
			wg.Done();
		}(key, pattern);
	}
	wg.Wait();
}

func readDir(path string, rec bool) {
	files, _ := os.ReadDir(path);
	for _, f := range files {
		tpath := filepath.Clean(path + string(os.PathSeparator) + f.Name());
		if f.IsDir() {
			if rec == true {
				fileList[tpath] = true;
				readDir(tpath, rec);
				continue;
			}
		}
		fileList[tpath] = true;
	}
}

// Returns the number of times pattern appears inside the given file
// TODO: Add multi pattern search
func searchFile(filename string, pattern string, reader io.Reader) int {
	//start := time.Now();
	//fmt.Println("reading file: ", filename);
	count := 0;
	buffer := make([]byte, BUFFER_SIZE);
	totalBytes := 0;
	totalMatches := 0;
	patternByte := []byte(pattern);
	for {
		bufferSize, err := reader.Read(buffer);
		if err != nil && err != io.EOF {
			return 0;
		}
		var position int;
		for {
			i := bytes.IndexByte(buffer[position:], '\n');
			if i == -1 || bufferSize == position {
				break;
			}
			if ignoreCase == true {
				if inLine := bytes.Count(bytes.ToLower(buffer[position:position+i]), bytes.ToLower(patternByte)); inLine > 0 {
					pm(filename + ":" + string(buffer[position:position+i]));
					totalMatches += inLine;
				}
			} else {
				if inLine := bytes.Count(buffer[position:position+i], patternByte); inLine > 0 {
					pm(filename + ":" +  string(buffer[position:position+i]));
					totalMatches += inLine;
				}
			}
			position += i + 1;
			count += 1;
		}
		totalBytes += position;
		if err == io.EOF {
			break;
		}

	}
	/*if totalMatches > 0 {
		fmt.Println("Total matches in ", filename, ":", totalMatches);
		fmt.Println("Total bytes in ", filename, ":", totalBytes);
		fmt.Println("Total time (ms) for ", filename, ":", time.Now().Sub(start).Milliseconds());
	}*/
	return count;
}

