package main

import (
	"bytes"
	"code.google.com/p/go-avltree/trunk"
	"io"
	"os"
	"strings"
	"sync/atomic"
	"stringsim/adjpair"
	"syscall"
)

type SimilarDir struct {
	// directory name
	dir string
	// pairs
	pairs adjpair.Pairs
	// similarity index
	six float64
}

// reverse comparsion
func (o SimilarDir) Compare(b avltree.Interface) int {
	if o.six > b.(SimilarDir).six {
		return -1
	}
	if o.six < b.(SimilarDir).six {
		return 1
	}
	return 0
}

type Cache struct {
	map_file *os.File
	buffer   []byte
}

type CacheReader struct {
	cache *Cache
	terminate int32
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

func NewCacheReader(cache *Cache) *CacheReader {
	return &CacheReader{cache: cache}
}

func (creader *CacheReader) isTerminating() bool {
    return atomic.LoadInt32(&creader.terminate) != 0
}

func (creader *CacheReader) setTerminate(value bool) {
    if value {
        atomic.StoreInt32(&creader.terminate, 1)
    } else {
        atomic.StoreInt32(&creader.terminate, 0)
    }
}

func (creader *CacheReader) Read(out chan *SimilarDir) error {
	buffer := bytes.NewBuffer(creader.cache.buffer)
	for {
		path, err := buffer.ReadString(0)
		if err == io.EOF {
			out <- nil
			close(out)
			break
		} else if err != nil {
			return err
		}
		out <- &SimilarDir{strings.TrimSpace(path), nil, 0.0}
		if creader.isTerminating() {
			close(out)
			break
		}
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
