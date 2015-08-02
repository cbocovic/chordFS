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

//Notify is part of the ChordApp interface and will update the
//application if its predecessor changes
func (fs *FileSystem) Notify(id []byte, myid []byte) string {
	//TODO: relocate all files in space
	return "tmp"

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
	ipaddr, err := chord.Lookup(key, addr)

	file, err := os.Open(path)
	defer file.Close()
	document := make([]byte, 4096)
	n, err := file.Read(document)
	checkError(err)
	if err != nil {
		return err
	}

	//create message to send to target ip
	fmt.Printf("Making store message with key=%s and doc=%s.\n", string(key[:32]), string(document[:n]))
	msg := getstoreMsg(key, document[:n-1])

	//send message TODO: check reply for errors
	fmt.Printf("Sending store message.\n")
	_, err = chord.Send(msg, ipaddr)

	return err

}

//Fetch will retrieve a file with key specified by key from the DHT and
//save it to path by contacting the node at addr
func Fetch(key [sha256.Size]byte, path string, addr string) error {
	ipaddr, err := chord.Lookup(key, addr)

	//create message to send to target ip
	fmt.Printf("Making fetch message with key=%s.\n", string(key[:32]))
	msg := getfetchMsg(key)

	fmt.Printf("Sending store message.\n")
	reply, err := chord.Send(msg, ipaddr)
	if err != nil {
		return err
	}

	reply = parseHeader(reply)
	document := parseDoc(reply)

	file, err := os.Create(path)
	checkError(err)
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
	file, err := os.Create(fmt.Sprintf("%s/%x", me.home, string(key)))
	checkError(err)
	_, err = file.Write(document)
	checkError(err)
	fmt.Printf("saved.\n")

	file.Close()

}

//loads a file from the node's home directory
func (me *FileSystem) load(key [sha256.Size]byte) []byte {

	document := make([]byte, 4096)
	filename := fmt.Sprintf("%x", key)
	file, err := os.Open(fmt.Sprintf("%s/%s", me.home, filename))
	defer file.Close()
	checkError(err)
	if err != nil {
		return document
	}
	n, err := file.Read(document)
	checkError(err)
	if err != nil {
		return document
	}

	return document[:n]
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
