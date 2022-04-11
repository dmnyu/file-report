package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/nyudlts/bytemath"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const version = "1.0.0"

var (
	extensions map[string]Extension
	inputDir   string
	outputFile string
	help       bool
)

func init() {
	extensions = make(map[string]Extension)
	flag.StringVar(&inputDir, "input-dir", "", "The Directory to walk")
	flag.StringVar(&outputFile, "output-file", "file-report.tsv", "/path/to/output/file")
	flag.BoolVar(&help, "help", false, "print this help screen")
}

func usage() {
	fmt.Println("\nusage: file-report [options]")
	fmt.Println("  options:")
	fmt.Println("    --input-dir /path/to/the/directory/to/walk \"Required\"")
	fmt.Printf("    --output-file /path/to/report [optional, default: file-report.tsv]\n\n")
}

func main() {

	//print help message if help flag is set
	if help == true {
		usage()
		os.Exit(0)
	}

	fmt.Printf("NYUDL File Report Tool v%s\n", version)
	fmt.Printf("* Parsing flags\n")
	//parse the flags
	flag.Parse()

	fmt.Printf("* Checking that input directory '%s' exists and is a directory\n", inputDir)
	//check that the directory exists and is a directory
	if err := rootExists(); err != nil {
		fmt.Printf("\n%s\n", err.Error())
		usage()
		os.Exit(1)
	}

	//walk the directory
	fmt.Printf("* Walking directory at %s\n", inputDir)
	err := filepath.Walk(inputDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			//fmt.Printf("Checking: %s\n", info.Name())
		} else {
			extension := strings.ToLower(filepath.Ext(info.Name()))
			if contains(extension) != true {
				extensions[extension] = Extension{Name: extension, Size: info.Size(), Count: 1}
			} else {
				tmpExt := extensions[extension]
				tmpExt.Count += 1
				tmpExt.Size += info.Size()
				extensions[extension] = tmpExt
			}
		}
		return nil
	})

	//check for any errors during walk
	fmt.Printf("* Checking for errors during walk\n")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(2)
	}

	//sort by size of files
	sortedExtensions := rankByWordCount(extensions)

	fmt.Printf("* Creating output tsv file\n")
	//create tsv file to write report to
	of, _ := os.Create(outputFile)
	defer of.Close()
	writer := bufio.NewWriter(of)
	writer.WriteString("Extension\tSize\tCount\tSize In Bytes\n")
	writer.Flush()

	fmt.Printf("* Writing output tsv to %s\n", outputFile)
	//create the output tsv
	var totalSize int64
	var totalCount int

	for _, entry := range sortedExtensions {
		if entry.Value.Size > 0 {
			totalSize += entry.Value.Size
			totalCount += entry.Value.Count
			size := bytemath.ConvertToHumanReadable(float64(entry.Value.Size))
			writer.WriteString(fmt.Sprintf("%s\t%s\t%d\t%d\n", entry.Value.Name, size, entry.Value.Count, entry.Value.Size))
			writer.Flush()
		}
	}

	//calculate human-readable size to total bytes
	humanSize := bytemath.ConvertToHumanReadable(float64(totalSize))

	//write totals to tsv file
	writer.WriteString(fmt.Sprintf("totals\t%s\t%d\t%d\n", humanSize, totalCount, totalSize))
	writer.Flush()

	//print quick summary
	fmt.Printf("* File-Report complete\n")
	fmt.Printf("  # files found: %d\n", totalCount)
	fmt.Printf("  total size of files %s\n", humanSize)
	fmt.Printf("* Exiting\n")
	os.Exit(0)
}

// support types and functions

type Extension struct {
	Name  string
	Count int
	Size  int64
}

type Pair struct {
	Key   string
	Value Extension
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value.Size < p[j].Value.Size }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func rootExists() error {
	if inputDir == "" {
		return fmt.Errorf("* [ERROR] inputDir is a required field")
	}
	fi, err := os.Stat(inputDir)
	if err == nil {

	} else if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("* [ERROR] location: '%s' does not exist", inputDir)

	} else {
		return err
	}

	if fi.IsDir() != true {
		return fmt.Errorf("* [ERROR] locaton: '%s' is not a direcotory", inputDir)
	}
	return nil
}

func contains(ext string) bool {
	for k, _ := range extensions {
		if k == ext {
			return true
		}
	}
	return false
}

func rankByWordCount(wordFrequencies map[string]Extension) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}
