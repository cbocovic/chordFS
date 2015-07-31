package fs

import (
	"crypto/sha256"
	"fmt"
	"github.com/cbocovic/chord"
	//"io/ioutil"
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

func Create(home string, addr string) *FileSystem {
	me := new(FileSystem)
	me.node = chord.Create(addr)
	me.home = home
	me.mirror = fmt.Sprintf("%s/mirrored", home)
	me.addr = addr

	me.node.Register(code, me)
	return me
}

func Join(home string, myaddr string, addr string) *FileSystem {
	me := new(FileSystem)
	me.node = chord.Join(myaddr, addr)
	me.home = home
	me.mirror = fmt.Sprintf("%s/mirrored", home)
	me.addr = myaddr
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
func (fs *FileSystem) Message(addr string, data string) string {
	//TODO: send to appropriate parsing function
	return "tmp"
}

//Store will store a file located at path in the DHT (under key) by
//contacting the node at addr
func Store(key [sha256.Size]byte, path string, addr string) error {
	//TODO: send file to appropriate node
	//do a lookup of the key
	targetip, err := chord.Lookup(key, addr)

	//create message to send to target ip
	msg := getstoreMsg(key, []byte(path))

	return err

}

//Fetch will retrieve a file with key specified by key from the DHT and
//save it to path by contacting the node at addr
func Fetch(key []byte, path string, addr string) {
	//TODO: find file from appropriate node

}

//saves the file to the node's home directory
func (me *FileSystem) save(key []byte, document []byte) {
	//TODO: save to file

}

//loads a file from the node's home directory
func (me *FileSystem) load(key []byte) string {
	//TODO: loads file
	return "tmp"
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
