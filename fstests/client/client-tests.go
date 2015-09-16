package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/cbocovic/chordFS"
	"io"
	"os"
	"os/exec"
	"time"
)

func main() {
	//First, join 1000 servers
	servers := make([]uint, 10)
	app := "screen"
	arg0 := "-dmS"
	arg1 := "1-tests"
	arg2 := "./tests"
	arg3 := "-start"
	arg4 := "1"
	cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4)
	cmd.Run()
	servers[0] = 1
	for i := 300; i < 309; i++ {
		app := "screen"
		arg0 := "-dmS"
		arg1 := fmt.Sprintf("%d-tests", i)
		arg2 := "./tests"
		arg3 := "-start"
		arg4 := fmt.Sprintf("%d", i)
		cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4)
		cmd.Run()
		servers[i-299] = uint(i)
	}
	fmt.Printf("Joined 10 servers. Sleeping now for 10 minutes.\n")

	//wait for stabilization
	time.Sleep(10 * time.Minute)
	log, _ := os.Create("results.log")
	log.Close()

	for j := 0; j < 1; j++ {
		fmt.Printf("Beginning run %d of 1.\n", j+1)
		log, _ = os.OpenFile("results.log", os.O_APPEND|os.O_WRONLY, 0666)
		log.Write([]byte(fmt.Sprintf("Beginning test %d of 1.\n", j+1)))
		log.Close()
		files := make([][sha256.Size]byte, 500)
		//put in documents
		for i := 0; i < 5; i++ {
			var key [sha256.Size]byte
			var addr uint32
			fmt.Printf("Planting file %d of 5.\n", i+1)
			rkey := make([]byte, sha256.Size)
			io.ReadFull(rand.Reader, rkey)
			copy(key[:], rkey)
			files[i] = key
			binary.Read(rand.Reader, binary.LittleEndian, &addr)
			addr = addr % 10
			raddr := servers[addr]
			fmt.Printf("random int: %d.\n", addr)
			low := (1 + raddr) % 256
			middle := ((1 + raddr) / 256) % 256
			high := ((1 + raddr) / (256 * 256)) % 256
			randaddr := fmt.Sprintf("127.%d.%d.%d:8888", high, middle, low)
			fs.Store(sha256.Sum256(key[:]), "index.php", randaddr)
		}

		//now join and leave 10 servers
		for i := 0; i < 5; i++ {
			fmt.Printf("Leaving server %d of 5.\n", i+1)
			var randomServer uint32
			var randomServer2 uint32
			binary.Read(rand.Reader, binary.LittleEndian, &randomServer)
			randomServer = randomServer % 10
			app := "screen"
			arg0 := "-X"
			arg1 := "-S"
			arg2 := fmt.Sprintf("%d-tests", servers[randomServer])
			arg3 := "quit"
			cmd := exec.Command(app, arg0, arg1, arg2, arg3)
			cmd.Run()

			fmt.Printf("Server %d just left.\n", servers[randomServer])
			fmt.Printf("Joining server %d of 5.\n", i+1)

			binary.Read(rand.Reader, binary.LittleEndian, &randomServer2)
			randomServer2 = randomServer2 % 10000000
			app = "screen"
			arg0 = "-dmS"
			arg1 = fmt.Sprintf("%d-tests", randomServer2)
			arg2 = "./tests"
			arg3 = "-start"
			arg4 = fmt.Sprintf("%d", randomServer2)
			cmd = exec.Command(app, arg0, arg1, arg2, arg3, arg4)
			cmd.Run()
			servers[uint(randomServer)] = uint(randomServer2)
			fmt.Printf("Server %d just joined.\n", servers[randomServer])

		}

		//wait for it to stabilize
		time.Sleep(10 * time.Minute)

		//recall files
		for i := 0; i < 5; i++ {
			var addr uint32
			fmt.Printf("Retrieving file %d of 5.\n", i+1)
			binary.Read(rand.Reader, binary.LittleEndian, &addr)
			addr = addr % 10
			raddr := servers[addr]
			low := (1 + raddr) % 256
			middle := ((1 + raddr) / 256) % 256
			high := ((1 + raddr) / (256 * 256)) % 256
			randaddr := fmt.Sprintf("127.%d.%d.%d:8888", high, middle, low)
			fs.Fetch(sha256.Sum256(files[i][:]), "fetched.txt", randaddr)
		}

		//delete all files
		fmt.Printf("Deleting files...")
		for _, server := range servers {
			dir, _ := os.Open(fmt.Sprintf("FS/%d", server))

			names, _ := dir.Readdirnames(0)
			dir.Close()

			//fmt.Printf("FS %s has %d files.\n", me.addr, len(names))
			for _, name := range names {
				//fmt.Printf("FS %s found file %s.\n", me.addr, name)
				os.Remove(name)
			}
		}
	}

}
