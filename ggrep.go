package main

import (
	"bufio"
	"fmt"
	"os"
	"io"
	"bytes"
	"strings"
	"sync"
	"time"
	"path/filepath"
	//"testing"
)

var wg sync.WaitGroup; // waitgroup

// Option Flags
var ignoreCase bool;
var recursive bool;

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

	temp := make([]string, 0);
	for _, f := range filenames {
		fmt.Println("looping over filesnames: ", f);
		if strings.HasSuffix(f, "/") {
			temp = append(temp, readDir(f, recursive)...);
		} else {
			temp = append(temp, f);
		}
	}

	//for _, f := range temp {
	//	fmt.Println("listing path:", f);
	//}
	filenames = temp;

	fmt.Println("Printing list of files to search through:");
	fmt.Println(filenames);
	fmt.Println("End of list of files to search through");

	// Start searching for patterns in files
	fmt.Println("Number of files to search:", len(filenames));
	for i := 0; i < len(filenames); i++ {
		if filenames[i] == "ggrep" {
			fmt.Println("Skipping the ggrep executable."); // Don't forget to remove this later.
			continue;
		}
		wg.Add(1);
		go func(fns string, p string) {
			file, _ := os.Open(fns);
			defer file.Close();
			searchFile(fns, p, bufio.NewReader(file));
			wg.Done();
		}(filenames[i], pattern);
	}
	wg.Wait();
	/*for {
		time.Sleep(0 * time.Second); // so that program doesnt reach end of main()
	}*/
}

// meant for later
/*func runSearch(filenames []string, pattern string) {
	fileCount := len(filenames);

	for i := 0; i < fileCount; i++ {
		if filenames[i] == "ggrep" {
			fmt.Println("Skipping the ggrep executable.");
			continue;
		}
		file, err := os.Open(filenames[i]);
		defer file.Close();
		if err != nil {
			fmt.Println("Error opening file: ", filenames[i]);
		}
		reader := bufio.NewReader(file);
		wg.Add(1);
		go func(f string, p string, r io.Reader) {
			searchFile(f, p, r);
			wg.Done();
		}(filenames[i], pattern, reader);
		wg.Wait(); // put oustide of the for loop
	}
}*/

/*NOTE:
	I didnt like the filepath.WalkDirFunc alternative to gathering a list of files
	Also since File and DirEntry types do not store the relative path to their respective file, I prepend the sub-directories to the current file.

	TODO: maybe just call the search function from right in here insteadof adding the relative path to a list, then looping over the list to search
		  find a way to min the append() calls
*/
func readDir(path string, rec bool) []string {
	files, _ := os.ReadDir(path);
	filenames := make([]string, 0); 
	for _, f := range files {
		if f.IsDir() {
			if rec == true {
				dirName := f.Name();
				tpath := filepath.Clean(path + string(os.PathSeparator) + dirName);
		        filenames = append(filenames, readDir(tpath, rec)...);
				continue;
			}
		}
		tpath := filepath.Clean(path + string(os.PathSeparator) + f.Name());
	    filenames = append(filenames, tpath);
	}
	return filenames;
}

// Returns the number of times pattern appears inside the given file
// TODO: Add multi pattern search
func searchFile(filename string, pattern string, reader io.Reader) int {
	start := time.Now();
	count := 0;
	buffer := make([]byte, bufio.MaxScanTokenSize);
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
					fmt.Println(filename, ":", string(buffer[position:position+i]));
					totalMatches += inLine;
				}
			} else {
				if inLine := bytes.Count(buffer[position:position+i], patternByte); inLine > 0 {
					fmt.Println(filename, ":", string(buffer[position:position+i]));
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
	if totalMatches > 0 {
		fmt.Println("Total matches in ", filename, ":", totalMatches);
		fmt.Println("Total bytes in ", filename, ":", totalBytes);
		fmt.Println("Total time (ms) for ", filename, ":", time.Now().Sub(start).Milliseconds());
	}
	return count;
}

