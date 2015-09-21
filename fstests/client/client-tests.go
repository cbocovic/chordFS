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
	arg0 := "-dmSL"
	arg1 := "1-tests"
	arg2 := "./tests"
	arg3 := "-start"
	arg4 := "1"
	cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4)
	cmd.Run()
	servers[0] = 1
	for i := 300; i < 309; i++ {
		app := "screen"
		arg0 := "-dmSL"
		arg1 := fmt.Sprintf("%d-tests", i)
		arg2 := "./tests"
		arg3 := "-start"
		arg4 := fmt.Sprintf("%d", i)
		cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4)
		cmd.Run()
		servers[i-299] = uint(i)
	}
	fmt.Printf("Joined 10 servers. Sleeping now for 5 minutes.\n")

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
			var newServer2 uint32
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

			//join to existing server
			existingAddr := servers[(randomServer+1)%10]
			low := (1 + existingAddr) % 256
			middle := ((1 + existingAddr) / 256) % 256
			high := ((1 + existingAddr) / (256 * 256)) % 256
			existingIpAddr := fmt.Sprintf("127.%d.%d.%d:8888", high, middle, low)

			//choose new ip address
			binary.Read(rand.Reader, binary.LittleEndian, &newServer2)
			newServer2 = newServer2 % 10000000
			app = "screen"
			arg0 = "-dmSL"
			arg1 = fmt.Sprintf("%d-tests", newServer2)
			arg2 = "./tests"
			arg3 = "-start"
			arg4 = fmt.Sprintf("%d", newServer2)
			arg5 := "-jointo"
			arg6 := existingIpAddr
			cmd = exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6)
			cmd.Run()
			servers[uint(randomServer)] = uint(newServer2)
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
			err := fs.Fetch(sha256.Sum256(files[i][:]), fmt.Sprintf("fetched-%d.txt", i), randaddr)
			if err != nil {
				fmt.Printf("Could not retrieve document. %s.\n", err.Error())
			}
		}

		//delete all files
		fmt.Printf("Deleting files...\n")
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
