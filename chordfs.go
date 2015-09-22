/* Package fs

is a file system application to be run on top of a Chord DHT.
*/
package chordfs

import (
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"github.com/cbocovic/chord"
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

	//testing purposes only
	malicious bool
	cache     map[string]bool
}

type FSError struct {
	address  string
	filename string
	Err      error
}

func (e *FSError) Error() string {
	if e != nil {
		return fmt.Sprintf("Failed to retrieve file %s:%s. Cause of failure: %s.", e.address, e.filename, e.Err)
	} else {
		return fmt.Sprintf("Failed to retrieve file %s:%s.", e.address, e.filename)
	}

}

//error checking function
func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, " chordFS: Fatal error: %s\n", err.Error())
	}
}

//Create will start a new DHT with a file system application and returns the FileSystem struct.
//The input home indicates the directory that will store distributed documents.
//The input addr specifies which address the application will listen on.
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
	//experimental
	me.malicious = false
	me.cache = make(map[string]bool)
	return me
}

//Join will create a Chord node with a file system application and join to an existing DHT
//specified by the address addr. It returns the resulting FileSystem struct.
//The input home indicates the directory that will store distributed documents.
//The input myaddr specifies which address the application will listen on.
func Join(home string, myaddr string, addr string) *FileSystem {
	var err error
	me := new(FileSystem)
	me.node, err = chord.Join(myaddr, addr)
	if err != nil {
		checkError(err)
		return nil
	}
	fmt.Printf("here\n")

	me.home = home
	me.mirror = fmt.Sprintf("%s/mirrored", home)
	me.addr = myaddr

	err = os.MkdirAll(me.home, 0755)
	err = os.MkdirAll(me.mirror, 0755)
	fmt.Printf("made directory %s.\n", me.home)
	if err != nil {
		checkError(err)
		return nil
	}
	me.node.Register(code, me)
	//experimental
	me.malicious = false
	me.cache = make(map[string]bool)
	return me
}

//Extend is similar to Join and Create, but instead takes as argument
//an already created ChordNode structure
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
//application by moving files if its predecessor changes
func (fs *FileSystem) Notify(id [sha256.Size]byte, myid [sha256.Size]byte, addr string) {
	fmt.Printf("Predecessor changed to %s.\n", addr)
	dir, err := os.Open(fs.home)
	defer dir.Close()
	checkError(err)
	if err != nil {
		return
	}

	names, err := dir.Readdirnames(0)
	checkError(err)
	if err != nil {
		return
	}

	for _, name := range names {
		var key [sha256.Size]byte
		decoded, err := base32.StdEncoding.DecodeString(name)
		if err != nil {
			//fmt.Printf("file was not encoded.\n")
			continue
		}
		copy(key[:], decoded[:sha256.Size])
		if chord.InRange(id, key, myid) {
			//transfer file.
			fs.cache[name] = true
			file, err := os.Open(fmt.Sprintf("%s/%s", fs.home, name))
			checkError(err)
			if err != nil {
				checkError(err)
			}
			defer file.Close()
			document := make([]byte, 4096)
			n, err := file.Read(document)
			checkError(err)
			if err != nil {
				checkError(err)
			}

			//create message to send to target ip
			msg := getstoreMsg(key, document[:n])

			//send message TODO: check reply for errors
			_, err = chord.Send(msg, addr)
			if err != nil {
				checkError(err)
			}

			fmt.Printf("Relocated file %s to node %s.\n", name, addr)

		}
	}

	//TODO: exchange mirrored files

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
	msg := getstoreMsg(key, document[:n])

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
	if err != nil {
		name := base32.StdEncoding.EncodeToString(key[:])
		err = &FSError{ipaddr, name, err}
		fmt.Printf("%s.\n", err.Error())
		return err
	}

	//create message to send to target ip
	msg := getfetchMsg(key)

	reply, err := chord.Send(msg, ipaddr)
	if err != nil {
		name := base32.StdEncoding.EncodeToString(key[:])
		err = &FSError{ipaddr, name, err}
		fmt.Printf("%s.\n", err.Error())
		return err
	}

	reply, err = parseHeader(reply)
	if err != nil {
		name := base32.StdEncoding.EncodeToString(key[:])
		err = &FSError{ipaddr, name, err}
		fmt.Printf("%s.\n", err.Error())
		return err
	}
	if reply == nil {
		name := base32.StdEncoding.EncodeToString(key[:])
		err = &FSError{ipaddr, name, err}
		fmt.Printf("%s.\n", err.Error())
		return err
	}

	document, err := parseDoc(reply)
	if err != nil {
		name := base32.StdEncoding.EncodeToString(key[:])
		err = &FSError{ipaddr, name, err}
		fmt.Printf("%s.\n", err.Error())
		return err
	}
	if document == nil {
		name := base32.StdEncoding.EncodeToString(key[:])
		err = &FSError{ipaddr, name, err}
		fmt.Printf("%s.\n", err.Error())
		return err
	}

	file, err := os.Create(path)
	checkError(err)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(document)
	checkError(err)
	if err != nil {
		return err
	}

	return err

}

//saves the file to the node's home directory
func (me *FileSystem) save(key []byte, document []byte) {
	name := base32.StdEncoding.EncodeToString(key)
	file, err := os.Create(fmt.Sprintf("%s/%s", me.home, name))
	checkError(err)
	_, err = file.Write(document)
	checkError(err)

	file.Close()

}

//loads a file from the node's home directory
func (me *FileSystem) load(key [sha256.Size]byte) ([]byte, error) {

	log, _ := os.OpenFile("results.log", os.O_APPEND|os.O_WRONLY, 0666) //experiments
	defer log.Close()

	document := make([]byte, 4096)
	name := base32.StdEncoding.EncodeToString(key[:])

	file, err := os.Open(fmt.Sprintf("%s/%s", me.home, name))

	defer file.Close()
	checkError(err)
	if err != nil {
		log.Write([]byte("File missing.\n"))
		return document, err
	}
	n, err := file.Read(document)
	checkError(err)
	if err != nil {
		log.Write([]byte("File missing.\n"))
		return document, err
	}

	//experiments
	if _, ok := me.cache[name]; ok {
		log.Write([]byte("Retrieved from cache.\n"))
	} else {
		log.Write([]byte("Retrieved from regular storage.\n"))
	}

	return document[:n], nil
}

//Allos the FileSystem to exit cleanly (removes all files for now)
func (fs *FileSystem) Finalize() {
	os.Remove(fs.home)
	fs.node.Finalize()

}

/** Printouts of information **/

func (fs *FileSystem) String() string {
	return fs.node.String()
}

func (fs *FileSystem) ShowFingers() string {
	return fs.node.ShowFingers()
}

func (fs *FileSystem) ShowSucc() string {
	return fs.node.ShowSucc()
}

//experimental purposes only
func (fs *FileSystem) MakeMalicious() {
	fs.malicious = true
}
