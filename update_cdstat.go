package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const maxStat = 100

func getStatFilePath() (p string, err error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	p = path.Clean(filepath.Join(usr.HomeDir, ".cdstat"))
	// fileInfo, err := os.Stat(path)
	// TODO later: maybe we should test for directory (and symlink)
	return p, nil
}

func closeMe(file io.Closer, fName string) {
	if err := file.Close(); err != nil {
		_ = fmt.Errorf("Can't close %s: %v", fName, err)
	}
}

type stat struct {
	count int
	date  int64
}

type stats map[string]stat

func updateStats(st stats, count int, date int64, path string) stats {
	if _, ok := st[path]; ok {
		st[path] = stat{count: st[path].count + count, date: date}
	} else {
		st[path] = stat{count: count, date: date}
	}
	return st
}

func readStats(input io.Reader) (st stats, err error) {
	st = make(stats)
	scanner := bufio.NewScanner(input)
	var ln int
	for scanner.Scan() {
		l := scanner.Text()
		items := strings.SplitN(l, " ", 3)
		if len(items) != 3 {
			msg := "%d:need count date path, got %d items"
			log.Printf(msg, ln, len(items))
		}
		count, err := strconv.Atoi(items[0])
		if err != nil {
			log.Printf("%d: can't parse %s", ln, items[0])
			continue
		}
		date, err := strconv.ParseInt(items[1], 10, 64)
		if err != nil {
			_ = fmt.Errorf("%d: can't parse %s", ln, items[1])
			continue
		}
		st = updateStats(st, count, date, items[2])
	}
	return st, nil
}

type pair struct {
	Key   string
	Value stat
}

type pairList []pair

func (p pairList) Len() int           { return len(p) }
func (p pairList) Less(i, j int) bool { return p[i].Value.count < p[j].Value.count }
func (p pairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func rankByStat(st stats) pairList {
	pl := make(pairList, len(st))
	i := 0
	for k, v := range st {
		pl[i] = pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

func main() {
	flag.Parse()
	newpath := flag.Arg(0)
	if newpath == "" {
		os.Exit(0)
	}
	now := time.Now().Unix()

	statFilePath, err := getStatFilePath()
	if err != nil {
		log.Fatalf("Can't get ~/.cdstat path: %v", err)
	}

	statFile, err := os.OpenFile(statFilePath, os.O_CREATE|os.O_RDWR, 0x700)
	if err != nil && (!os.IsNotExist(err)) {
		log.Fatalf("Can't open %s: %v", statFilePath, err)
	}
	defer closeMe(statFile, statFilePath)

	st, err := readStats(statFile)
	if err != nil {
		log.Fatalf("can't read stats: %v", err)
	}
	st = updateStats(st, 1, now, newpath)

	pairs := rankByStat(st)
	if len(pairs) > maxStat {
		pairs = pairs[:maxStat]
	}

	var pos int64
	_, err = statFile.Seek(pos, 0)
	if err != nil {
		log.Fatalf("Can't seek to begin of statfile: %v", err)
	}

	outBuf := bufio.NewWriter(statFile)
	for _, p := range pairs {
		path, stat := p.Key, p.Value
		toWrite := fmt.Sprintf("%d %d %s\n", stat.count, stat.date, path)
		written, err := outBuf.Write([]byte(toWrite))
		pos += int64(written)
		if err != nil {
			_ = fmt.Errorf("can't write: %s: %v", toWrite, err)
			_ = outBuf.Flush()
			os.Exit(4)
		}
	}
	_ = outBuf.Flush()
	err = statFile.Truncate(pos)
	if err != nil {
		log.Fatalf("unable to truncate: %v", err)
	}
	_ = statFile.Sync()
}
