package fs

import (
	"crypto/sha256"
	"fmt"
	"github.com/cbocovic/chord"
	//"io"
	"os"
)

const (
	code byte = 2
)

type FileSystem struct {
	home   string
	mirror string

	node *chord.ChordNode
	addr string
}

//error checking function
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
	}
}

func Create(home string, addr string) *FileSystem {
	me := new(FileSystem)
	me.node = chord.Create(addr)
	if me.node == nil {
		return nil
	}
	me.home = home
	me.mirror = fmt.Sprintf("%s/mirrored", home)
	me.addr = addr
	//make directories
	err := os.MkdirAll(me.home, 0755)
	err = os.MkdirAll(me.mirror, 0755)
	fmt.Printf("made directory %s.\n", me.home)
	if err != nil {
		checkError(err)
		return nil
	}

	me.node.Register(code, me)
	return me
}

func Join(home string, myaddr string, addr string) *FileSystem {
	me := new(FileSystem)
	me.node = chord.Join(myaddr, addr)
	if me.node == nil {
		return nil
	}

	me.home = home
	me.mirror = fmt.Sprintf("%s/mirrored", home)
	me.addr = myaddr

	err := os.MkdirAll(me.home, 0755)
	err = os.MkdirAll(me.mirror, 0755)
	fmt.Printf("made directory %s.\n", me.home)
	if err != nil {
		checkError(err)
		return nil
	}
	me.node.Register(code, me)
	return me
}

//Extend is similar to Join and Create, but instead takes as argument
//a ChordNode structure
func Extend(home string, addr string, node *chord.ChordNode) *FileSystem {
	me := new(FileSystem)
	me.node = node

	me.home = home
	me.mirror = fmt.Sprintf("%s/mirrored", home)
	me.addr = addr

	err := os.MkdirAll(me.home, 0755)
	err = os.MkdirAll(me.mirror, 0755)
	fmt.Printf("made directory %s.\n", me.home)
	if err != nil {
		checkError(err)
		return nil
	}
	me.node.Register(code, me)
	return me
}

//Notify is part of the ChordApp interface and will update the
//application if its predecessor changes
func (me *FileSystem) Notify(id [sha256.Size]byte, myid [sha256.Size]byte) {
	fmt.Printf("FS %s notified!\n", me.addr)
	dir, err := os.Open(me.home)
	checkError(err)
	if err != nil {
		return
	}

	names, err := dir.Readdirnames(0)
	checkError(err)
	if err != nil {
		return
	}

	fmt.Printf("FS %s has %d files.\n", me.addr, len(names))
	for _, name := range names {
		fmt.Printf("FS %s found file %x.\n", me.addr, []byte(name))
		var key [sha256.Size]byte
		if len(name) < sha256.Size {
			continue
		}
		copy(key[:], []byte(name)[:sha256.Size])
		if chord.InRange(id, key, myid) {
			fmt.Printf("Relocating file... ")
			err := Store(key, fmt.Sprintf("%s/%s", me.home, name), me.addr)
			if err != nil {
				fmt.Printf("error: ")
				checkError(err)
			} else {
				fmt.Printf("done.\n")
			}
		}
	}

}

//Message is part of the ChordApp interface and will allow chord
//to forward messages to the application
func (fs *FileSystem) Message(data []byte) []byte {
	return fs.parseMessage(data)
}

//Store will store a file located at path in the DHT (under key) by
//contacting the node at addr
func Store(key [sha256.Size]byte, path string, addr string) error {
	//do a lookup of the key
	fmt.Printf("storing... \n")
	ipaddr, err := chord.Lookup(key, addr)
	fmt.Printf("belongs to %s.\n", ipaddr)

	file, err := os.Open(path)
	checkError(err)
	if err != nil {
		fmt.Printf("error here (0)\n")
		return err
	}
	defer file.Close()
	document := make([]byte, 4096)
	n, err := file.Read(document)
	checkError(err)
	if err != nil {
		fmt.Printf("error here (1)\n")
		return err
	}

	//create message to send to target ip
	msg := getstoreMsg(key, document[:n-1])

	//send message TODO: check reply for errors
	_, err = chord.Send(msg, ipaddr)
	if err != nil {
		fmt.Printf("error here (2)\n")
	}

	return err

}

//Fetch will retrieve a file with key specified by key from the DHT and
//save it to path by contacting the node at addr
func Fetch(key [sha256.Size]byte, path string, addr string) error {
	ipaddr, err := chord.Lookup(key, addr)

	//create message to send to target ip
	msg := getfetchMsg(key)

	reply, err := chord.Send(msg, ipaddr)
	if err != nil {
		return err
	}

	reply, err = parseHeader(reply)
	if err != nil {
		return err
	}
	document := parseDoc(reply)

	file, err := os.Create(path)
	checkError(err)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Printf("Writing to file... ")
	_, err = file.Write(document)
	checkError(err)
	if err != nil {
		return err
	}
	fmt.Printf("done.\n")

	return err

}

//saves the file to the node's home directory
func (me *FileSystem) save(key []byte, document []byte) {
	fmt.Printf("saving... ")
	file, err := os.Create(fmt.Sprintf("%s/%s", me.home, string(key)))
	checkError(err)
	_, err = file.Write(document)
	checkError(err)
	fmt.Printf("saved.\n")

	file.Close()

}

//loads a file from the node's home directory
func (me *FileSystem) load(key [sha256.Size]byte) ([]byte, error) {

	document := make([]byte, 4096)
	file, err := os.Open(fmt.Sprintf("%s/%s", me.home, string(key[:sha256.Size])))
	defer file.Close()
	checkError(err)
	if err != nil {
		return document, err
	}
	n, err := file.Read(document)
	checkError(err)
	if err != nil {
		return document, err
	}

	return document[:n], nil
}

//exits cleanly (removes all files for now)
func (fs *FileSystem) Finalize() {
	fs.node.Finalize()

}

/** Printouts of information **/

func (fs *FileSystem) Info() string {
	return fs.node.Info()
}

func (fs *FileSystem) ShowFingers() string {
	return fs.node.ShowFingers()
}

func (fs *FileSystem) ShowSucc() string {
	return fs.node.ShowSucc()
}
