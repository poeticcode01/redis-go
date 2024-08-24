package main

import (
	"fmt"
	"net"
	"os"
)

func  handleClient(conn net.Conn){
	// ensure we close the connection after we're done
	defer conn.Close()

	// Read data
	for {
		buf := make([]byte,1024)
		n,err := conn.Read(buf)

		if err != nil{
			fmt.Println("Error reading data from  the client",err.Error())
			return
		}
		fmt.Println("Received data",string(buf[:n]))


		message := []byte("+PONG\r\n")
		n,err = conn.Write(message)

		if err != nil{
			fmt.Println("Error sending message to the  client",err.Error())
			return
		}
		fmt.Printf("send %d bytes",n)

	}
	



}

func main() {
	
	listener, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	// ensure to stop the tcp server when the program exits
	defer listener.Close()

	fmt.Println("Server is listening on port 6379")
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
