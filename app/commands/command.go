package commands

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/hashtable"
)

const (
	ClrfDelimeter = "\r\n"
)

var DEFAULTROLE string = "master"
var MASTER_REPLID string = "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb"
var MASTER_REPL_OFFSET int = 0
var REPLICATION_SERVER_PORT string = ""
var RELICATION_COUNT int = 0

func clrfSplit(str string) []string {
	command_slice := strings.Split(str, ClrfDelimeter)
	return command_slice

}

type CliCommand interface {
	Run([]string) (string, error)
}

func Execute(input_buf string) (string, error) {
	command_slice := Decode(input_buf)
	fmt.Println("Command slice is  ", command_slice)

	name := strings.ToLower(command_slice[0])
	var cmd CliCommand

	switch name {
	case "ping":
		cmd = &Ping{}
	case "echo":
		cmd = &Echo{}
	case "set":
		cmd = &Set{}
	case "get":
		cmd = &Get{}
	case "info":
		cmd = &Info{}
	case "replconf":
		cmd = &ReplConf{}
	case "psync":
		cmd = &Psync{}
	case "wait":
		cmd = &Wait{}
	}

	return (cmd).Run(command_slice[1:])

}

type Info struct{}
type Echo struct{}
type Ping struct{}
type Set struct{}
type Get struct{}
type ReplConf struct{}
type Psync struct{}
type Wait struct{}

type Type interface {
	Encode() string
}
type SimpleString struct {
	Content string
}

type BulkString struct {
	Content *string
}

func (*Wait) Run(input []string) (string, error) {
	msg := fmt.Sprintf(":%d\r\n", RELICATION_COUNT)
	return msg, nil
}

func (*Psync) Run(input []string) (string, error) {
	var respType Type = &SimpleString{Content: fmt.Sprintf("FULLRESYNC %s %d", MASTER_REPLID, MASTER_REPL_OFFSET)}
	return respType.Encode(), nil

}

func (*ReplConf) Run(input []string) (string, error) {
	// REPLCONF GETACK *
	REPLICATION_SERVER_PORT = input[1]
	fmt.Println("relication command", input)
	var respType Type = &SimpleString{Content: "OK"}
	return respType.Encode(), nil

}

func (*Info) Run(input []string) (string, error) {
	info_argument := strings.ToLower(input[0])
	var resp string = ""

	switch info_argument {
	case "replication":
		return replication()

	}
	return resp, nil
}

func replication() (string, error) {
	content := fmt.Sprintf("role:%s\nmaster_replid:%s\nmaster_repl_offset:%d", DEFAULTROLE, MASTER_REPLID, MASTER_REPL_OFFSET)
	var respType Type = &BulkString{Content: &content}
	return respType.Encode(), nil

}

func (*Echo) Run(input []string) (string, error) {
	if len(input) == 1 {
		bulk := &BulkString{Content: &input[0]}
		return bulk.Encode(), nil
	}
	arr := make([]*BulkString, 0, len(input))
	for _, v := range input {
		arr = append(arr, &BulkString{Content: &v})
	}
	return EncodeArray(arr, true), nil
}

func CreateMessage(message_list []string) (string, error) {

	arr := make([]*BulkString, 0, len(message_list))
	for _, message := range message_list {
		arr = append(arr, &BulkString{Content: &message})
	}

	// fmt.Println(fmt.Sprintf("message %v", *arr[0]))
	return EncodeArray(arr, false), nil
}

func (*Ping) Run(_ []string) (string, error) {
	var respType Type = &SimpleString{Content: "PONG"}
	return respType.Encode(), nil
}

func (*Set) Run(data_slice []string) (string, error) {
	fmt.Println("Data slice is ", data_slice)
	key, value := data_slice[0], data_slice[1]
	var px string = ""
	if len(data_slice) > 2 {
		_, px = data_slice[2], data_slice[3]
	}
	fmt.Printf("Request to Set Key %s to Value %s\n", key, value)
	cache := hashtable.GetCache()
	err := cache.Set(key, value, px)
	// return_val, err := cache.Get(key)
	if err != nil {
		fmt.Printf("Error setting up key %s\n", key)
	}
	fmt.Printf("Key %s Set to %s\n", key, value)
	var respType Type = &SimpleString{Content: "OK"}
	return respType.Encode(), nil

}

func (*Get) Run(data_slice []string) (string, error) {
	key := data_slice[0]
	cache := hashtable.GetCache()
	return_val, err := cache.Get(key)

	if err != nil {
		fmt.Printf("Key %s not exists or Expired \n", key)
		var nullContent *string = nil
		nullBs := BulkString{Content: nullContent}
		return nullBs.Encode(), nil
	}
	fmt.Printf("Returning Value %s for  Key %s \n", return_val, key)
	bulk := &BulkString{Content: &return_val}
	return bulk.Encode(), nil

}

func (t *BulkString) Encode() string {
	if t.Content == nil {
		// Return null bulk string if Content is nil
		return fmt.Sprintf("$-1%s", ClrfDelimeter)
	}
	str := *t.Content
	length := len(str)
	return fmt.Sprintf("$%d%s%s%s", length, ClrfDelimeter, str, ClrfDelimeter)
}

func EncodeArray[T Type](arr []T, extra bool) string {
	length := len(arr)
	arrMark := fmt.Sprintf("*%d%s", length, ClrfDelimeter)
	buf := bytes.NewBuffer([]byte(arrMark))
	if extra {
		for _, v := range arr {
			encoded := fmt.Sprintf("%v%s", v.Encode(), ClrfDelimeter)
			buf.WriteString(encoded)
		}

	} else {
		for _, v := range arr {
			encoded := fmt.Sprintf("%v", v.Encode())
			buf.WriteString(encoded)
		}

	}

	return buf.String()
}

func (t *SimpleString) Encode() string {
	str := t.Content
	return fmt.Sprintf("+%s%s", str, ClrfDelimeter)
}

func Decode(str string) []string {
	split := clrfSplit(str)
	fmt.Println("Split input is", split)
	result := make([]string, 0, len(split))
	for i := 0; i < len(split); i++ {
		switch split[i][0] {
		case '$':
			result = append(result, split[i+1])
			// fmt.Println("current index in $", i, split[i])
		case '+':
			result = append(result, split[i][1:])
			// fmt.Println("current index in +", i, split[i])
		default:
			// fmt.Println("current index in default", i, split[i])
			continue
		}

	}
	return result
}
