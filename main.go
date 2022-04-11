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

var (
	extensions map[string]Extension
	inputDir   string
)

type Pair struct {
	Key   string
	Value Extension
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value.Size < p[j].Value.Size }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func init() {
	extensions = make(map[string]Extension)
	flag.StringVar(&inputDir, "dir", "", "The Directory to walk")
}

type Extension struct {
	Name  string
	Count int
	Size  int64
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

func RootExists() error {
	fi, err := os.Stat(inputDir)
	if err == nil {

	} else if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s does not exist", inputDir)

	} else {
		return err
	}

	if fi.IsDir() != true {
		return fmt.Errorf("input location is not a direcotory")
	}
	return nil
}

func main() {
	flag.Parse()
	if err := RootExists(); err != nil {
		panic(err)
	}

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

	if err != nil {
		fmt.Println(err.Error())
	}

	sortedExtensions := rankByWordCount(extensions)

	of, _ := os.Create("file-report.tsv")
	defer of.Close()
	writer := bufio.NewWriter(of)

	for _, entry := range sortedExtensions {
		if entry.Value.Size > 0 {
			size := bytemath.ConvertToHumanReadable(float64(entry.Value.Size))
			writer.WriteString(fmt.Sprintf("%s\t%s\t%d\t%d\n", entry.Value.Name, size, entry.Value.Count, entry.Value.Size))
			writer.Flush()
		}
	}

}
