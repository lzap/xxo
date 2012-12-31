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

// the last similarity element indicating workers to stop
var LAST_DIR SimilarDir = SimilarDir{"", nil, 0.0}

func PreparePairs(in chan *SimilarDir, out chan *SimilarDir) {
	for {
		dir := <-in
		if dir == nil {
			out <- nil
			break
		}
		dir.pairs = adjpair.NewPairsFromFilepath(dir.dir)
		out <- dir
	}
}

func ComputeSimilarities(search_pairs adjpair.Pairs, in chan *SimilarDir, out chan *SimilarDir) {
	for {
		dir := <-in
		if dir == nil {
			out <- nil
			break
		}
		dir.six = search_pairs.Match(dir.pairs)
		out <- dir
	}
}

func ReadKeys(stdscr Window, out chan Key) {
	for {
		ch := stdscr.GetChar()
		out <- ch
		if KeyString(ch) == "q" {
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
	win.ScrollOk(true)

	printmenu(&win, result_tree, active)

	// key reading channel
	key_ch := make(chan Key)
	go ReadKeys(stdscr, key_ch)

	// searching channels
	read_ch := make(chan *SimilarDir, CHANNEL_BUFFER)
	prepare_ch := make(chan *SimilarDir, CHANNEL_BUFFER)
	compute_ch := make(chan *SimilarDir, CHANNEL_BUFFER)

	// stop channel
	stop_ch := make(chan int)

	// refresh screen channel
	tick_ch := time.Tick(time.Millisecond * 250)
	needs_refresh := true

	for {
		select {
		case <-tick_ch:
			if needs_refresh {
				printmenu(&win, result_tree, active)

				stdscr.Print(0, 1, "Query: "+query)
				stdscr.ClearToEOL()
				stdscr.Print(rows-1, 1, "Results: "+string(result_tree.Len()))
				stdscr.ClearToEOL()
				stdscr.Refresh()
				needs_refresh = false
			}
		case ch := <-key_ch:
			switch KeyString(ch) {
			case "esc":
				return
			case "up":
				if active > 0 {
					active -= 1
				}
				printmenu(&win, result_tree, active)
			case "down":
				if active < result_tree.Len()-1 {
					active += 1
				}
				printmenu(&win, result_tree, active)
			case "enter":
				//stdscr.Print(23, 0, "Choice #%d: %s selected", active, result_tree[active])
				stdscr.ClearToEOL()
				stdscr.Refresh()
			default:
				// some other key - restart search
				query = query + KeyString(ch)
				// send stop signal
				//stop_ch <- 0
				// immediately start new routines for new computation
				search_pairs := adjpair.NewPairsFromString(query)
				go cache.Read(stop_ch, read_ch)
				go PreparePairs(read_ch, prepare_ch)
				go ComputeSimilarities(search_pairs, prepare_ch, compute_ch)
			}
		case simdir := <-compute_ch:
			if simdir == nil {
				// last filepath - set clear flag
				//result_tree.Clear()
				needs_refresh = true
			} else if simdir.six > SIMILARITY_CUT {
				result_tree.Add(*simdir)
				needs_refresh = true
			}
		}
	}
}

func printmenu(w *Window, result_tree *avltree.ObjectTree, active int) {
	_, max_width := w.Maxyx()
	max_width = max_width - 5
	y, x := 1, 2
	w.Box(0, 0)
	i := 0
	for v := range result_tree.Iter() {
		label := v.(SimilarDir).dir
		if len(label) > max_width && max_width > 0 {
			label = label[0:max_width]
		}
		if i == active {
			w.AttrOn(A_REVERSE)
			w.Print(y+i, x, label)
			w.AttrOff(A_REVERSE)
		} else {
			w.Print(y+i, x, label)
		}
		i += 1
	}
	w.Refresh()
}
