package main

import "fmt"
import "path/filepath"
import "os"
import "io/fs"
import "strings"
import "bufio"
import "github.com/nyudlts/bytemath"
import "sort"

var exts map[string]int64

func contains(ext string) bool {
	for k, _ := range exts {
		if k == ext {
			return true
		}
	}
	return false
}

type Pair struct {
	Key   string
	Value int64
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func rankByWordCount(wordFrequencies map[string]int64) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

func main() {
	exts = make(map[string]int64)
	inputDir := "/mnt/ACMBornDigital/Collections/fa/mss/610/"
	err := filepath.Walk(inputDir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			//fmt.Printf("Checking: %s\n", info.Name())
		} else {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if contains(ext) != true {
				exts[ext] = info.Size()
			} else {
				exts[ext] = exts[ext] + info.Size()
			}
		}

		return nil
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	sortedExts := rankByWordCount(exts)

	of, _ := os.Create("ayers.tsv")
	defer of.Close()
	writer := bufio.NewWriter(of)

	for _, entry := range sortedExts {
		if entry.Value > 0 {
			//fmt.Printf("%s\t%d\n", k,v)
			size := bytemath.ConvertToHumanReadable(float64(entry.Value))
			writer.WriteString(fmt.Sprintf("%s\t%s\n", entry.Key, size))
			writer.Flush()
		}
	}
}
