package main

import (
	. "code.google.com/p/goncurses"
	"fmt"
	"os"
  "stringsim/adjpair"
)

const (
	HEIGHT = 10
	WIDTH  = 30
)

const SIMILARITY_CUT float64 = 0.00005
const CHANNEL_BUFFER = 1024
var LAST_DIR SimilarDir = SimilarDir{"", nil, 0.0}

func PreparePairs(in chan *SimilarDir, out chan *SimilarDir) {
  for {
    dir := <-in
    dir.pairs = adjpair.NewPairsFromFilepath(dir.dir)
    out <- dir
    if dir == &LAST_DIR {
      break
    }
  }
}

func ComputeSimilarities(search_pairs adjpair.Pairs, in chan *SimilarDir, out chan *SimilarDir) {
  for {
    dir := <-in
    dir.six = search_pairs.Match(dir.pairs)
    out <- dir
    if dir == &LAST_DIR {
      break
    }
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
	result_slice := []string{}

	stdscr, _ := Init()
	defer End()

	Raw(true)
	Echo(false)
	Cursor(0)
	stdscr.Clear()
	stdscr.Keypad(true)

	rows, cols := stdscr.Maxyx()
	win, _ := NewWindow(rows - 2, cols - 1, 1, 0)
	win.Keypad(true)

	printmenu(&win, result_slice, active)

	// key reading channel
  key_ch := make(chan Key)
	go ReadKeys(stdscr, key_ch)

	// searching channels
  read_ch := make(chan *SimilarDir, CHANNEL_BUFFER)
  prepare_ch := make(chan *SimilarDir, CHANNEL_BUFFER)
  compute_ch := make(chan *SimilarDir, CHANNEL_BUFFER)

	for {
		select {
		case ch := <- key_ch:
			switch KeyString(ch) {
			case "q":
				return
			case "up":
				if active == 0 {
					active = len(result_slice) - 1
				} else {
					active -= 1
				}
			case "down":
				if active == len(result_slice)-1 {
					active = 0
				} else {
					active += 1
				}
			case "enter":
				stdscr.Print(23, 0, "Choice #%d: %s selected", active, result_slice[active])
				stdscr.ClearToEOL()
				stdscr.Refresh()
			default:
				// some other key - restart search
				query = query + KeyString(ch)
  			search_pairs := adjpair.NewPairsFromString(query)
				go cache.Read(read_ch)
				go PreparePairs(read_ch, prepare_ch)
				go ComputeSimilarities(search_pairs, prepare_ch, compute_ch)
			}
		case simdir := <-compute_ch:
			if simdir.six > SIMILARITY_CUT {
				result_slice = append(result_slice, simdir.dir)
			}
			if simdir == &LAST_DIR {
				// last filepath - set clear flag
				result_slice = []string{}
			}
		}

		stdscr.Print(0, 0, " Query: " + query)
		stdscr.ClearToEOL()
		stdscr.Refresh()

		printmenu(&win, result_slice, active)
	}
}

func printmenu(w *Window, result_slice []string, active int) {
	y, x := 2, 2
	w.Box(0, 0)
	for i, s := range result_slice {
		if i == active {
			w.AttrOn(A_REVERSE)
			w.Print(y+i, x, s)
			w.AttrOff(A_REVERSE)
		} else {
			w.Print(y+i, x, s)
		}
	}
	w.Refresh()
}
