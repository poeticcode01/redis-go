package commands

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	ClrfDelimeter = "\r\n"
)

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
	}

	return (cmd).Run(command_slice[1:])

}

type Echo struct{}

func (*Echo) Run(input []string) (string, error) {
	if len(input) == 1 {
		bulk := &BulkString{Content: input[0]}
		return bulk.Encode(), nil
	}
	arr := make([]*BulkString, 0, len(input))
	for _, v := range input {
		arr = append(arr, &BulkString{Content: v})
	}
	return EncodeArray(arr), nil
}

type Ping struct{}

func (*Ping) Run(_ []string) (string, error) {
	var respType Type = &SimpleString{Content: "PONG"}
	return respType.Encode(), nil
}

type Type interface {
	Encode() string
}

type BulkString struct {
	Content string
}

func (t *BulkString) Encode() string {
	str := t.Content
	length := len(str)
	return fmt.Sprintf("$%d%s%s%s", length, ClrfDelimeter, str, ClrfDelimeter)
}

func EncodeArray[T Type](arr []T) string {
	length := len(arr)
	arrMark := fmt.Sprintf("*%d%s", length, ClrfDelimeter)
	buf := bytes.NewBuffer([]byte(arrMark))
	for _, v := range arr {
		encoded := fmt.Sprintf("%v%s", v.Encode(), ClrfDelimeter)
		buf.WriteString(encoded)
	}
	return buf.String()
}

type SimpleString struct {
	Content string
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
