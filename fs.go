package fs

import (
	"fmt"
	"githum.com/cbocovic/chord"
	"io/ioutil"
)

type FileSystem struct {
	home   string
	mirror string

	node chord.ChordNode
	addr string
}

func Create(home string, addr string) *FileSystem {
	me := new(FileSystem)
	me.node = chord.Create(addr)
	me.home = home
	me.mirror = fmt.Sprintf("%s/mirrored", home)
	me.addr = addr
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
func (me *FileSystem) Notify(id []byte, me []byte) {

}

//Message is part of the ChordApp interface and will allow chord
//to forward messages to the application
func (me *FileSystem) Message(addr string, data string) {

}

//Store will store a file located at path in the DHT (under key) by
//contacting the node at addr
func Store(key []byte, path string, addr string) {

}

//Fetch will retrieve a file with key specified by key from the DHT and
//save it to path by contacting the node at addr
func Fetch(key []byte, path string, addr string) {

}

//saves the file to the node's home directory
func (me *FileSystem) save(key []byte, path string) {

}

//loads a file from the node's home directory
func (me *FileSystem) load(key []byte) string {
}

//exits cleanly (removes all files for now)
func (me *FileSystem) finalize() {

}
