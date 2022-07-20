package main

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"io/ioutil"
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
	fmt.Println("filenames: ", filenames);

	if len(filenames) == 1 && filenames[0] == "." {

		//append(filenames,
		files, _ := ioutil.ReadDir("./");
		for _, f := range files {
			if f.IsDir() {
				fmt.Println("ignoring: ", f.Name());
				continue;
			}
			filenames = append(filenames, f.Name());
			fmt.Println("appended: ", f.Name());
		}
		filenames = filenames[1:];
	}


	files, _ := ioutil.ReadDir("./");
	for i, f := range files {
		if f.IsDir() {
			continue;
		}
		fmt.Println(i, " number:", f.Name());
    }


//	if len(filenames) == 1 && filenames[0] == "." {
		//files, _ := ioutil.ReadDir("./");
		//fmt.Println("printing out ./");
		//fmt.Println(files.Name());
		fmt.Println("finished printing");
//	}

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
			if ignoreCase == true {
				stringBuffer = strings.ToLower(stringBuffer);
				pattern = strings.ToLower(pattern);
			}
			if inLine := strings.Count(stringBuffer, pattern); inLine  > 0 {
				fmt.Println(filename, ":", stringBuffer);
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
	fmt.Println("Total matches in ", filename, ":", totalMatches);
	fmt.Println("Total bytes in ", filename, ":", totalBytes);
	fmt.Println("Total time for ", filename, ":", time.Now().Sub(start));
	return count;
}

