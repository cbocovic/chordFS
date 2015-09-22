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
	servers := make([]uint, 1000)
	app := "screen"
	arg0 := "-dmSL"
	arg1 := "1-tests"
	arg2 := "./tests"
	arg3 := "-start"
	arg4 := "1"
	cmd := exec.Command(app, arg0, arg1, arg2, arg3, arg4)
	cmd.Run()
	servers[0] = 1
	for i := 300; i < 1299; i++ {
		//join to existing server
		existingAddr := servers[i-300]
		low := (1 + existingAddr) % 256
		middle := ((1 + existingAddr) / 256) % 256
		high := ((1 + existingAddr) / (256 * 256)) % 256
		existingIpAddr := fmt.Sprintf("127.%d.%d.%d:8888", high, middle, low)

		app := "screen"
		arg0 := "-dmSL"
		arg1 := fmt.Sprintf("%d-tests", i)
		arg2 := "./tests"
		arg3 := "-start"
		arg4 := fmt.Sprintf("%d", i)
		arg5 := "-jointo"
		arg6 := existingIpAddr
		cmd = exec.Command(app, arg0, arg1, arg2, arg3, arg4, arg5, arg6)
		cmd.Run()
		servers[i-299] = uint(i)
		time.Sleep(1 * time.Minute)
	}
	fmt.Printf("Joined 1000 servers. Sleeping now for 1 hour.\n")

	//wait for stabilization
	time.Sleep(1 * time.Hour)
	log, _ := os.Create("results.log")
	log.Close()

	for j := 0; j < 100; j++ {
		fmt.Printf("Beginning run %d of 1.\n", j+1)
		log, _ = os.OpenFile("results.log", os.O_APPEND|os.O_WRONLY, 0666)
		log.Write([]byte(fmt.Sprintf("Beginning test %d of 1.\n", j+1)))
		log.Close()
		files := make([][sha256.Size]byte, 500)
		//put in documents

		fmt.Printf("Beginning to plant files.\n")
		for i := 0; i < 500; i++ {
			var key [sha256.Size]byte
			var addr uint32
			fmt.Printf("\rPlanting file %d of 500.\n", i+1)
			rkey := make([]byte, sha256.Size)
			io.ReadFull(rand.Reader, rkey)
			copy(key[:], rkey)
			files[i] = key
			binary.Read(rand.Reader, binary.LittleEndian, &addr)
			addr = addr % 1000
			raddr := servers[addr]
			low := (1 + raddr) % 256
			middle := ((1 + raddr) / 256) % 256
			high := ((1 + raddr) / (256 * 256)) % 256
			randaddr := fmt.Sprintf("127.%d.%d.%d:8888", high, middle, low)
			chordfs.Store(sha256.Sum256(key[:]), "index.php", randaddr)
			time.Sleep(10 * time.Second)
		}

		//now join and leave 10 servers
		fmt.Printf("Beginning join-and-leave step.\n")
		for i := 0; i < 10; i++ {
			fmt.Printf("\rLeaving server %d of 10.\n", i+1)
			var randomServer uint32
			var newServer2 uint32
			binary.Read(rand.Reader, binary.LittleEndian, &randomServer)
			randomServer = randomServer % 1000
			app := "screen"
			arg0 := "-X"
			arg1 := "-S"
			arg2 := fmt.Sprintf("%d-tests", servers[randomServer])
			arg3 := "quit"
			cmd := exec.Command(app, arg0, arg1, arg2, arg3)
			cmd.Run()

			//join to existing server
			existingAddr := servers[(randomServer+1)%1000]
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

		}

		//wait for it to stabilize
		time.Sleep(1 * time.Hour)

		//recall files
		fmt.Printf("Retrieving files.\n")
		for i := 0; i < 500; i++ {
			var addr uint32
			fmt.Printf("\rRetrieving file %d of 500.\n", i+1)
			binary.Read(rand.Reader, binary.LittleEndian, &addr)
			addr = addr % 1000
			raddr := servers[addr]
			low := (1 + raddr) % 256
			middle := ((1 + raddr) / 256) % 256
			high := ((1 + raddr) / (256 * 256)) % 256
			randaddr := fmt.Sprintf("127.%d.%d.%d:8888", high, middle, low)
			err := chordfs.Fetch(sha256.Sum256(files[i][:]), fmt.Sprintf("fetched.txt", i), randaddr)
			if err != nil {
				fmt.Printf("Could not retrieve document.\n")
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
