package main

import (
	"code.google.com/p/go-avltree/trunk"
	. "code.google.com/p/goncurses"
	"fmt"
	"os"
	"stringsim/adjpair"
	"time"
)

const (
	HEIGHT = 10
	WIDTH  = 30
)

// similarity bellow this threashold are cut off
const SIMILARITY_CUT float64 = 0.00005

// maximum results to hold in the AVL-tree
const MAX_RESULTS = 1024

// buffer size  for all processing channels (do not set it too high)
const CHANNEL_BUFFER = 32

func PreparePairs(in chan *SimilarDir, out chan *SimilarDir) {
	defer close(out) 
	for dir := range in { 
		dir.pairs = adjpair.NewPairsFromFilepath(dir.dir)
		out <- dir
	} 
}

func ComputeSimilarities(search_pairs adjpair.Pairs, in chan *SimilarDir, out chan *SimilarDir) {
	defer close(out) 
	for dir := range in { 
		dir.six = search_pairs.Match(dir.pairs)
		out <- dir
	} 
}

func ReadKeys(stdscr Window, out chan Key) {
	defer close(out) 
	for {
		ch := stdscr.GetChar()
		out <- ch
		if ch == 27 || ch == 'q' {
			break
		}
	}
}

func main() {
	// prepare cache
	cache, err := OpenCache("var")
	if err != nil {
		fmt.Println("error opening file ", err)
		os.Exit(1)
	}
	defer cache.Close()

	var active int
	var query = ""
	result_tree := avltree.NewObjectTree(0)

	stdscr, _ := Init()
	defer End()

	Raw(true)
	Echo(false)
	Cursor(0)
	stdscr.Clear()
	stdscr.Keypad(true)

	rows, cols := stdscr.Maxyx()
	win, _ := NewWindow(rows-3, cols-1, 1, 0)
	win.Keypad(true)

	// TODO subwindow?

	scroll_lines := 0

	// key reading channel
	key_ch := make(chan Key)
	go ReadKeys(stdscr, key_ch)

	// searching channels
	var read_ch, prepare_ch, compute_ch chan *SimilarDir
	
	// reader instance (the first worker in the chain)
	var reader *CacheReader

	// refresh screen channel
	tick_ch := time.Tick(time.Millisecond * 250)
	refresh := true
	clear := false
	message := ""
	refresh_count := 0

	for {
		select {
		case <-tick_ch:
			if refresh {
				RedrawResults(&win, result_tree, active, scroll_lines)

				stdscr.Print(0, 2, "Query: "+query)
				stdscr.ClearToEOL()
				message = fmt.Sprintf("read=%d, prepare=%d, compute=%d", len(read_ch), len(prepare_ch), len(compute_ch))
				stdscr.Print(rows-2, 2, fmt.Sprintf("Results: %d %d %s %d", result_tree.Len(), scroll_lines, message, refresh_count))
				stdscr.ClearToEOL()
				stdscr.Refresh()
				refresh = false
				refresh_count += 1
			}
		case ch := <-key_ch:
			switch ch {
			case 27:
				return
			case 'q':
				return
			case KEY_PAGEUP:
				scroll_lines -= 10
				if scroll_lines < 0 {
					scroll_lines = 0
				}
				// refresh immediately
				RedrawResults(&win, result_tree, active, scroll_lines)
			case KEY_PAGEDOWN:
				scroll_lines += 10
				// refresh immediately
				RedrawResults(&win, result_tree, active, scroll_lines)
			case KEY_UP:
				if active > 0 {
					active -= 1
				}
				// refresh immediately
				RedrawResults(&win, result_tree, active, scroll_lines)
			case KEY_DOWN:
				if active < result_tree.Len()-1 {
					active += 1
				}
				// refresh immediately
				RedrawResults(&win, result_tree, active, scroll_lines)
			case KEY_ENTER:
				//stdscr.Print(23, 0, "Choice #%d: %s selected", active, result_tree[active])
				stdscr.ClearToEOL()
				stdscr.Refresh()
			default:
				// some other key - restart search
				query = query + KeyString(ch)
				search_pairs := adjpair.NewPairsFromString(query)
				// send stop signal to the old reader
				if reader != nil {
					reader.setTerminate(true)
				}
				// immediately start new goroutines for new computation
				read_ch = make(chan *SimilarDir, CHANNEL_BUFFER)
				prepare_ch = make(chan *SimilarDir, CHANNEL_BUFFER)
				compute_ch = make(chan *SimilarDir, CHANNEL_BUFFER)
				reader = NewCacheReader(cache)
				go reader.Read(read_ch)
				go PreparePairs(read_ch, prepare_ch)
				go ComputeSimilarities(search_pairs, prepare_ch, compute_ch)
			}
		case simdir, ok := <-compute_ch:
			if ! ok {
				// input channel was closed - set clear flag
				clear = true
				refresh = true
				// and set compute channel to nil channel
				compute_ch = nil
			} else if simdir.six > SIMILARITY_CUT {
				// some results are being sent
				if clear {
					result_tree.Clear()
					clear = false
				}
				result_tree.Add(*simdir)
				refresh = true
			}
		}
	}
}

func RedrawResults(w *Window, result_tree *avltree.ObjectTree, active int, scroll_lines int) {
	_, max_width := w.Maxyx()
	max_width = max_width - 5
	y, x := 1, 2
	i := -1
	for v := range result_tree.Iter() {
		i += 1
		if i <= scroll_lines {
			continue
		}
		label := v.(SimilarDir).dir
		if len(label) > max_width && max_width > 0 {
			label = label[0:max_width]
		}
		if i == active {
			w.AttrOn(A_REVERSE)
			w.Print(y, x, label)
			w.AttrOff(A_REVERSE)
		} else {
			w.Print(y, x, label)
		}
		w.ClearToEOL()
		y += 1
	}
	w.Box(0, 0)
	w.Refresh()
}
