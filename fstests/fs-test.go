package main

import (
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"github.com/cbocovic/chord"
	"github.com/cbocovic/chordFS"
	"io"
	//"runtime"
	"time"
)

func main() {
	//runtime.GOMAXPROCS(runtime.NumCPU())

	var startaddr string

	//set up flags
	numPtr := flag.Int("num", 1, "the size of the DHT you wish to test")
	startPtr := flag.Int("start", 1, "ipaddr to start from")
	jointoPtr := flag.String("jointo", "127.0.0.2:8888", "existing ipaddr to join to")

	flag.Parse()
	num := *numPtr
	start := *startPtr
	jointo := *jointoPtr

	low := (1 + start) % 256
	middle := ((1 + start) / 256) % 256
	high := ((1 + start) / (256 * 256)) % 256
	startaddr = fmt.Sprintf("127.%d.%d.%d:8888", high, middle, low)

	fmt.Printf("WOOO Joining %d server starting at %s!\n", 1, startaddr)

	list := make([]*chordfs.FileSystem, num)
	if start == 1 {
		me := new(chordfs.FileSystem)
		me = chordfs.Create(fmt.Sprintf("FS/%s", startaddr), startaddr)
		list[0] = me
	} else {
		me := new(chordfs.FileSystem)
		me = chordfs.Join(fmt.Sprintf("FS/%s", startaddr), startaddr, jointo)
		list[0] = me
	}

	for i := 1; i < num; i++ {
		//join node to network or start a new network
		time.Sleep(time.Second)
		node := new(chordfs.FileSystem)
		low := (1 + start + i) % 256
		middle := ((1 + start + i) / 256) % 256
		high := ((1 + start + i) / (256 * 256)) % 256
		addr := fmt.Sprintf("127.%d.%d.%d:8888", high, middle, low)

		fmt.Printf("Joining %d server starting at %s!\n", 1, addr)
		node = chordfs.Join(fmt.Sprintf("FS/%s", addr), addr, startaddr)
		list[i] = node
		fmt.Printf("Joined server: %s.\n", addr)
	}
	//block until receive input
Loop:
	for {
		var cmd string
		var index int
		var ipaddr string
		_, err := fmt.Scan(&cmd)
		switch {
		case cmd == "info":
			//print out successors and predecessors
			fmt.Printf("Node\t\t Successor\t\t Predecessor\n")
			for _, node := range list {
				fmt.Printf("%s\n", node.String())
			}
		case cmd == "fingers":
			//print out finger table
			fmt.Printf("Enter index of desired node: ")
			fmt.Scan(&index)
			if index >= 0 && index < len(list) {
				node := list[index]
				fmt.Printf("\n%s", node.ShowFingers())
			}
		case cmd == "succ":
			//print out successor list
			fmt.Printf("Enter index of desired node: ")
			fmt.Scan(&index)
			if index >= 0 && index < len(list) {
				node := list[index]
				fmt.Printf("\n%s", node.ShowSucc())
			}
		case err == io.EOF:
			break Loop
		case cmd == "lookup":
			var key [sha256.Size]byte
			fmt.Printf("Enter ipaddr of valid node: ")
			fmt.Scan(&ipaddr)
			rkey := make([]byte, sha256.Size)
			io.ReadFull(rand.Reader, rkey)
			copy(key[:], rkey)
			addr, err := chord.Lookup(key, ipaddr)
			if err != nil {
				fmt.Printf("Fatal error.%s\n.", err.Error())
			} else {
				fmt.Printf("Found key at %s.\n", addr)
			}
		}

	}
	for _, node := range list {
		node.Finalize()
	}

}
