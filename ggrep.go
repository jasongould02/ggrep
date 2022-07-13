package main

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"bytes"
	"strings"
	"time"
)

// Option Flags
var ignoreCase bool;

func main() {
	args := os.Args[1:];

	lastOptionIndex := 0;
	for i := 0; i < len(args); i++ {
		// search for elements that start with hypen, then check for that type of command
		// -ig = ignore case matches
		// -mp = patterns are in a [] list delimited by commas
		// -r = search all files in current folder and sub-folders (recursive search)
		// if filename is . then search all files in the current folder
		switch args[i] {
			case "-ig": // ignore cases when searching
				lastOptionIndex = i;

				ignoreCase = true;
			case "-mp": // parse for multiple patterns
				lastOptionIndex = i;
				fmt.Println("-mp is not implemented yet");
			case "-r": // search for pattern(s) recursively
				lastOptionIndex = i;
				fmt.Println("-r is not implemented yet");
			default: // not a recognized argument option
		}

	}

	pattern := args[(lastOptionIndex+1)];
	filenames := args[(lastOptionIndex + 2):];
	for i := 0; i < len(filenames); i++ {
		file, err := os.OpenFile(filenames[i], os.O_RDONLY, os.ModePerm);
		if err != nil {
			fmt.Println("Error opening file: ", filenames[i]);
		}
		reader := bufio.NewReader(file);
		go searchFile(filenames[i], pattern, reader);
		defer file.Close();
	}
	for {
		time.Sleep(0 * time.Second); // so that program doesnt reach end of main()
	}
}

func recursiveSearch(rootFolder string, pattern string, reader io.Reader) int {
	return 0;
}

// Returns the number of times pattern appears inside the given file
func searchFile(filename string, pattern string, reader io.Reader) int {
	start := time.Now();
	fmt.Println("parsing file ", filename);
	count := 0;
	buffer := make([]byte, bufio.MaxScanTokenSize);
	totalBytes := 0;
	totalMatches := 0;
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
			stringBuffer := strings.ToLower(string(buffer[position:position+i]));
			//fmt.Println("%d : %s")
			if ignoreCase == true {
				stringBuffer = strings.ToLower(stringBuffer);
				pattern = strings.ToLower(pattern);
			}
			if inLine := strings.Count(stringBuffer, pattern); inLine  > 0 {
				fmt.Println(inLine, " : ", stringBuffer);
				totalMatches += inLine;
			}

			position += i + 1;

			count += 1;
		}
		totalBytes += position;
		if err == io.EOF {
			break;
		}

	}
	fmt.Println("Total matches in getLineCount()", totalMatches);
	fmt.Println("Total Bytes:", totalBytes);
	fmt.Println("Total Time:", time.Now().Sub(start));
	return count;
}

