package main

import (
  "os"
  "syscall"
  "strings"
  "bytes"
  "io"
  "stringsim/adjpair"
)

type SimilarDir struct {
	// directory name
  dir string
	// pairs
  pairs adjpair.Pairs
	// similarity index
  six float64
}

type Cache struct {
	map_file *os.File
	buffer []byte
}

func OpenCache(filename string) (*Cache, error) {
	cache := new(Cache)
  file, err := os.Open(filename)
  if err != nil {
    return nil, err
  }
	cache.map_file = file
  stat, err := cache.map_file.Stat()
  if err != nil {
    return nil, err
  }
  buffer, err := syscall.Mmap((int)(cache.map_file.Fd()), 0, (int)(stat.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
  if err != nil {
    return nil, err
  }
	cache.buffer = buffer
	return cache, nil
}

func (cache *Cache) Read(out chan *SimilarDir) error {
  buffer := bytes.NewBuffer(cache.buffer)
  for {
    path, err := buffer.ReadString(0)
    if err == io.EOF {
      out <-&LAST_DIR
      break
    } else if err != nil {
      return err
    }
    dir := SimilarDir{strings.TrimSpace(path), nil, 0.0}
    out <-&dir
  }
	return nil
}

func (cache *Cache) Close() error {
	err := syscall.Munmap(cache.buffer)
  if err != nil {
    return err
  }
	err = cache.map_file.Close()
  if err != nil {
    return err
  }
	return nil
}

/*
// simple implementation
func ReadCache2(filename string, out chan *SimilarDir) error {
  f, err := os.Open(filename)
  defer f.Close()
  r := bufio.NewReader(f)
  for {
    path, err := r.ReadString(0)
    if err == io.EOF {
      out <-&LAST_DIR
      break
    } else if err != nil {
      return err
    }
    dir := SimilarDir{strings.TrimSpace(path), nil, 0.0}
    out <-&dir
  }
  return nil
}
*/
