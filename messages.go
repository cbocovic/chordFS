package chordfs

import (
	"crypto/sha256"
	"fmt"
	"github.com/cbocovic/chordFS/internal"
	"github.com/golang/protobuf/proto"
	"log"
)

func getstoreMsg(key [sha256.Size]byte, document []byte) []byte {

	msg := new(fsMsgs.NetworkMessage)
	msg.Proto = proto.Uint32(2)
	appMsg := new(fsMsgs.AppMessage)
	fsMsg := new(fsMsgs.AppMessage_FileSystemMessage)
	command := fsMsgs.AppMessage_Command(fsMsgs.AppMessage_Command_value["STORE"])
	fsMsg.Cmd = &command
	storeMsg := new(fsMsgs.StoreMessage)
	storeMsg.Key = proto.String(string(key[:sha256.Size]))
	storeMsg.Document = proto.String(string(document))
	fsMsg.Smsg = storeMsg
	appMsg.Msg = fsMsg

	appdata, err := proto.Marshal(appMsg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	msg.Msg = proto.String(string(appdata))

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	return data
}

func getfetchMsg(key [sha256.Size]byte) []byte {

	msg := new(fsMsgs.NetworkMessage)
	msg.Proto = proto.Uint32(2)
	appMsg := new(fsMsgs.AppMessage)
	fsMsg := new(fsMsgs.AppMessage_FileSystemMessage)
	command := fsMsgs.AppMessage_Command(fsMsgs.AppMessage_Command_value["FETCH"])
	fsMsg.Cmd = &command
	fetchMsg := new(fsMsgs.FetchMessage)
	fetchMsg.Key = proto.String(string(key[:sha256.Size]))
	fsMsg.Fmsg = fetchMsg
	appMsg.Msg = fsMsg

	appdata, err := proto.Marshal(appMsg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	msg.Msg = proto.String(string(appdata))

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	return data
}

func nullMsg() []byte {
	msg := new(fsMsgs.NetworkMessage)
	msg.Proto = proto.Uint32(2)

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	return data
}

func (fs *FileSystem) parseMessage(data []byte) []byte {

	msg := new(fsMsgs.AppMessage)

	err := proto.Unmarshal(data, msg)
	checkError(err)
	if err != nil {
		fmt.Printf("Uh oh in fs parse message of node %s\n", fs.addr)
		return nullMsg()
	}

	fsmsg := msg.GetMsg()
	cmd := int32(fsmsg.GetCmd())
	switch {
	case cmd == fsMsgs.AppMessage_Command_value["FETCH"]:
		fmsg := fsmsg.GetFmsg()
		var key [sha256.Size]byte
		copy(key[:], []byte(fmsg.GetKey()))
		doc, err := fs.load(key)
		if err != nil {
			return nullMsg()
		}
		//experiment purposes
		if fs.malicious {
			fmt.Printf("Muahahaha malicious server.\n")
			return nullMsg()
		}
		return getstoreMsg(key, doc)
	case cmd == fsMsgs.AppMessage_Command_value["STORE"]:
		smsg := fsmsg.GetSmsg()
		key := []byte(smsg.GetKey())
		doc := []byte(smsg.GetDocument())
		fs.save(key, doc)
		return nullMsg()
	case cmd == fsMsgs.AppMessage_Command_value["MIRROR"]:
		return nullMsg()
	}
	fmt.Printf("No matching commands.\n")
	return nullMsg()
}

func parseDoc(data []byte) ([]byte, error) {

	msg := new(fsMsgs.AppMessage)

	err := proto.Unmarshal(data, msg)
	checkError(err)
	if err != nil {
		fmt.Printf("Uh oh in fs parse doc of node.\n")
		return nil, err
	}

	fsmsg := msg.GetMsg()
	smsg := fsmsg.GetSmsg()
	doc := smsg.GetDocument()
	return []byte(doc), nil

}

//parseHeader strips the Chord overlay layer off the message
func parseHeader(data []byte) ([]byte, error) {

	msg := new(fsMsgs.NetworkMessage)

	err := proto.Unmarshal(data, msg)
	checkError(err)
	if err != nil {
		fmt.Printf("Uh oh in header parse message of node.\n")
		return make([]byte, 0), err
	}

	protocol := msg.GetProto()
	if byte(protocol) != code {
		fmt.Printf("Uh oh in header parse message of node.\n")
		return make([]byte, 0), err
	}

	return []byte(msg.GetMsg()), nil
}
