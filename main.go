package main

import (
  "runtime"
  "fmt"
  "os"
  "syscall"
  "strings"
  "bytes"
  "bufio"
  "io"
  "stringsim/adjpair"
)

type SimilarDir struct {
  dir string // directory name
  pairs adjpair.Pairs // pairs
  six float64 // similarity index
}

var LAST_DIR SimilarDir = SimilarDir{"", nil, 0.0}
const CHANNEL_BUFFER = 1024

func ReadCache2(filename string, out chan *SimilarDir) error {
  f, err := os.Open(filename)
  if err != nil {
    fmt.Println("error opening file ", err)
    os.Exit(1)
  }
  defer f.Close()
  r := bufio.NewReader(f)
  for {
    path, err := r.ReadString(0)
    if err == io.EOF {
      out <- &LAST_DIR
      break
    } else if err != nil {
      return err
    }
    dir := SimilarDir{strings.TrimSpace(path), nil, 0.0}
    out <- &dir
  }
  return nil
}

func ReadCache(filename string, out chan *SimilarDir) error {
  // TODO optimize open/close (encaptulate)
  map_file, err := os.Open(filename)
  if err != nil {
    return err
  }
  stat, err := map_file.Stat()
  if err != nil {
    return err
  }
  b, err := syscall.Mmap((int)(map_file.Fd()), 0, (int)(stat.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
  if err != nil {
    return err
  }

  buffer := bytes.NewBuffer(b)
  for {
    path, err := buffer.ReadString(0)
    if err == io.EOF {
      out <- &LAST_DIR
      break
    } else if err != nil {
      return err
    }
    dir := SimilarDir{strings.TrimSpace(path), nil, 0.0}
    out <- &dir
  }

  // defer?
  err = syscall.Munmap(b)
  if err != nil {
    return err
  }
  err = map_file.Close()
  if err != nil {
    return err
  }
  return nil
}

func PreparePairs(in chan *SimilarDir, out chan *SimilarDir) {
  for {
    dir := <- in
    dir.pairs = adjpair.NewPairsFromString(dir.dir)
    out <- dir
    if dir == &LAST_DIR {
      break
    }
  }
}

func ComputeSimilarities(search_pairs adjpair.Pairs, in chan *SimilarDir, out chan *SimilarDir) {
  for {
    dir := <- in
    dir.six = search_pairs.Match(dir.pairs)
    out <- dir
    if dir == &LAST_DIR {
      break
    }
  }
}

func Search(searchstr string, max_results int) {
  search_pairs := adjpair.NewPairsFromString(searchstr)
  read_ch := make(chan *SimilarDir, CHANNEL_BUFFER)
  prepare_ch := make(chan *SimilarDir, CHANNEL_BUFFER)
  compute_ch := make(chan *SimilarDir, CHANNEL_BUFFER)
  go ReadCache("var", read_ch)
  go PreparePairs(read_ch, prepare_ch)
  go ComputeSimilarities(search_pairs, prepare_ch, compute_ch)
  for {
    simdir := <- compute_ch
    fmt.Println(simdir.dir, simdir.six)
    if len(simdir.dir) == 0 {
      break
    }
  }
}

func main() {
  runtime.GOMAXPROCS(4)
  Search("conf", 30)
}
