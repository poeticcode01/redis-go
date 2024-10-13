package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/commands"
)

const (
	ClrfDelimeter = "\r\n"
)

var slaves = []net.Conn{}
var master net.Conn

func HandeCommandsExtra(input_buf string, conn net.Conn) {
	command_slice := commands.Decode(input_buf)
	fmt.Println("Command slice is  ", command_slice)

	name := strings.ToLower(command_slice[0])

	switch name {
	case "psync":
		handle_psync(input_buf, conn)
	case "set":
		message, _ := commands.CreateMessage(command_slice)
		CallReplica(message)

	}

}

func handleClient(conn net.Conn) {
	// ensure we close the connection after we're done
	defer conn.Close()

	// Read data
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Error reading data from  the client", err.Error())
			return
		}

		input_buf := string(buf)
		output, err := commands.Execute(input_buf)
		fmt.Println("Output string is", output)

		if err != nil {
			fmt.Println("Error executing command: ", err.Error())
		}

		// message := []byte("+PONG\r\n")
		_, err = conn.Write([]byte(output))

		if err != nil {
			fmt.Println("Error sending message to the  client", err.Error())
			return
		}
		HandeCommandsExtra(input_buf, conn)
		// fmt.Printf("send %d bytes", n)

	}

}

func SyncWithMaster(conn net.Conn) {

	fmt.Println("***Start accepting the sync commands from master***")
	byte_read_so_far := 0
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)

		if err != nil {
			fmt.Println("Error reading data from  the master", err.Error())
			return
		}

		input_buf := string(buf[:n])

		fmt.Println("Received sync commands from master", input_buf)
		byte_len := len(input_buf)
		fmt.Println("Input bytes length", byte_len)
		split := strings.Split(input_buf, ClrfDelimeter)
		split_len := len(split)
		// fmt.Println("lenght of sync command is ", split_len)
		// fmt.Println("sync comand Split input is", split)
		// if split_len < 7 {
		// 	fmt.Println("Received commands not suppoted yet")
		// 	continue
		// }
		command_slice := []string{}
		// strt := 0

		// for strt < split_len-1 {
		// 	temp := split[strt : strt+7]
		// 	temp_command := strings.Join(temp, commands.ClrfDelimeter)
		// 	command_slice = append(command_slice, temp_command)
		// 	strt += 7
		// }
		const (
			SetCommand      = "set"
			PingCommand     = "ping"
			ReplconfCommand = "replconf"
		)

		prv := []string{}
		for ind := range split {
			// fmt.Println("here", ind, split[ind])
			if ind+1 >= split_len {
				break
			}
			cur_name := strings.ToLower(split[ind+1])
			// fmt.Println("cur_name", cur_name)

			// Create a helper function to handle command processing
			processCommand := func() {
				temp_command := strings.Join(prv, commands.ClrfDelimeter)
				fmt.Println("Temp command", temp_command, len(temp_command))
				if len(temp_command) > 7 {
					temp_prv := strings.ToLower(prv[1])
					if temp_prv == SetCommand || temp_prv == PingCommand || temp_prv == ReplconfCommand {
						command_slice = append(command_slice, temp_command)
					} else {
						fmt.Println("Unwanted commands")
					}

				} else {
					fmt.Println("Unwanted commands")
				}

				prv = []string{split[ind]}
				// fmt.Println("New beginning", prv, command_slice)
			}

			switch cur_name {
			case SetCommand, PingCommand, ReplconfCommand:
				processCommand()
			default:
				prv = append(prv, split[ind])
			}
			// fmt.Println("Running prev", prv)
		}
		temp_command := strings.Join(prv, commands.ClrfDelimeter)
		fmt.Println("Temp command", temp_command, len(temp_command))
		if len(temp_command) > 7 {
			command_slice = append(command_slice, temp_command)
		} else {
			fmt.Println("Unwanted commands")
		}
		// command_slice = append(command_slice, temp_command)
		fmt.Println("Command Slice", command_slice, len(command_slice))

		for _, sync_command := range command_slice {
			_, err := commands.Execute(sync_command)
			fmt.Println("Current command length is ", len(sync_command))
			// fmt.Println("Output string is", output)

			if err != nil {
				fmt.Println("Error executing command: ", err.Error())
			}
			// fmt.Println("Executed sync command", sync_command)

			temp_slice := commands.Decode(sync_command)
			name := strings.ToLower(temp_slice[0])
			// fmt.Println("Replication command from master", name)
			if name == "replconf" {
				// fmt.Println("Handle ACK")
				// *3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$1\r\n0\r\n.

				message_slice := []string{"REPLCONF", "ACK", strconv.Itoa(byte_read_so_far)}
				output, _ := commands.CreateMessage(message_slice)
				_, err = conn.Write([]byte(output))

				if err != nil {
					fmt.Println("Error sending ACK message to the  Master", err.Error())
					return
				}
				// fmt.Println("Successfullt sent ACK")

			}
			if name == PingCommand {
				byte_read_so_far += 14

			} else if name == ReplconfCommand {
				byte_read_so_far += 37
			} else {

				fmt.Println("Temp slice is  ", temp_slice)
				output, _ := commands.CreateMessage(temp_slice)
				fmt.Println("Resp len for set command is ", len(output))
				byte_read_so_far += len(output)

			}

			fmt.Println("Updated byte length is", byte_read_so_far)

		}
		// if len(command_slice) >= 1 {
		// 	fmt.Println("Bytes read so far", byte_read_so_far)
		// 	fmt.Println("current byte size is", byte_len)
		// 	byte_read_so_far += byte_len
		// 	fmt.Println("Updated byte size is", byte_read_so_far)
		// }

		fmt.Println("Synced the request command from master!")
	}

}

func handle_psync(input_buf string, conn net.Conn) {
	command_slice := commands.Decode(input_buf)
	fmt.Println("Command slice is  ", command_slice)

	remoteAddr := conn.RemoteAddr().String()
	fmt.Println("Received connection from:", remoteAddr)
	remote_address := strings.Split(remoteAddr, ":")
	slaves = append(slaves, conn)

	commands.REPLICATION_SERVER_PORT = remote_address[len(remote_address)-1]

	// name := strings.ToLower(command_slice[0])

	sync_data := "$87\r\nREDIS0011\xfa\tredis-ver\x057.2.0\xfa\nredis-bits\xc0@\xfa\x05ctime\xc2m\x08\xbce\xfa\x08used-mem\xc2\xb0\xc4\x10\x00\xfa\x08aof-base\xc0\x00\xff\xf0n;\xfe\xc0\xffZ"
	_, err := conn.Write([]byte(sync_data))
	fmt.Println("Sent empty RDB file")
	if err != nil {
		fmt.Println("Error sending empty RDB to the client  ", err.Error())
		return
	}
}

func CallReplica(message string) {
	// fmt.Println("Replica port", commands.REPLICATION_SERVER_PORT)
	// conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", replica_server, commands.REPLICATION_SERVER_PORT))
	// if err != nil {
	// 	fmt.Println("Error in Calling replica:", err)
	// 	os.Exit(1)
	// }
	// defer conn.Close()
	for _, slave := range slaves {
		_, err := slave.Write([]byte(message))

		if err != nil {
			fmt.Println("Error sending message to Replica", err)
			os.Exit(1)

		}

	}

	fmt.Println("Message Sent to replica :", message)

}

func HandShake(master_host string, master_port string, listening_port string) {

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", master_host, master_port))
	if err != nil {
		fmt.Println("Error in handshaking:", err)
		os.Exit(1)
	}
	// defer conn.Close()

	message_slice := [][]string{{"PING"},
		{"REPLCONF", "listening-port", listening_port},
		{"REPLCONF", "capa", "psync2"},
		{"PSYNC", "?", "-1"},
	}

	for _, message := range message_slice {

		bulk_message, _ := commands.CreateMessage(message)
		// fmt.Println("Bulk message is :", bulk_message)

		_, err = conn.Write([]byte(bulk_message))
		if err != nil {
			fmt.Println("Error sending message over Handshake:", err)
			os.Exit(1)

		}
		fmt.Println("Sent message:", message)

		// Wait for the server's response
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Error receiving response:", err)
			return
		}

		// Print the server's response
		fmt.Println("Server response:", response)

	}
	fmt.Println("Successfully HandShake done and Message sent")

	master = conn
	go SyncWithMaster(master)

}

func main() {
	port := flag.String("port", "6379", "server port")
	replicaof := flag.String("replicaof", "master", "decalre server as slave or master")

	// Parse the flags
	flag.Parse()

	server_port := fmt.Sprintf("0.0.0.0:%s", *port)

	if *replicaof != "master" {
		commands.DEFAULTROLE = "slave"
		master_info := strings.Split(*replicaof, " ")
		master_host := master_info[0]
		master_port := master_info[1]
		fmt.Println("Master info", master_host, master_port)
		go HandShake(master_host, master_port, *port)

	}

	// listener, err := net.Listen("tcp", "0.0.0.0:6379")
	listener, err := net.Listen("tcp", server_port)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	// ensure to stop the tcp server when the program exits
	defer listener.Close()

	fmt.Println("Server is listening on port:", *port)
	// Block until we receive an incoming connection

	for {

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)

		}
		// Hnadle client connection
		go handleClient(conn)

	}

}
