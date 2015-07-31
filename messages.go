package fs

import (
	//"fmt"
	"crypto/sha256"
	"github.com/golang/protobuf/proto"
	"log"
)

func getstoreMsg(key [sha256.Size]byte, document []byte) []byte {

	msg := new(NetworkMessage)
	msg.Proto = proto.Uint32(2)
	fsMsg := new(NetworkMessage_FileSystemMessage)
	command := NetworkMessage_Command(NetworkMessage_Command_value["STORE"])
	fsMsg.Cmd = &command
	storeMsg := new(StoreMessage)
	storeMsg.Key = proto.String(string(key[:sha256.Size]))
	storeMsg.Document = proto.String(string(document))
	fsMsg.Msg = storeMsg
	msg.Msg = fsMsg

	data, err := proto.Marshal(msg)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}

	return data
}
