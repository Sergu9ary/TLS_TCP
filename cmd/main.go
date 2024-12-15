package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

func InitTcpServer(address, privateKey, certificate string) {
	logFile, err := os.OpenFile("sslkey.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Failed to open key log file:", err)
		return
	}
	defer logFile.Close()

	certPair, err := tls.LoadX509KeyPair(certificate, privateKey)
	if err != nil {
		fmt.Println("Error loading TLS credentials:", err)
		return
	}

	listenerConfig := tls.Config{Certificates: []tls.Certificate{certPair}, InsecureSkipVerify: true, KeyLogWriter: logFile}
	tlsListener, err := tls.Listen("tcp", address+":8443", &listenerConfig)
	if err != nil {
		fmt.Println("Unable to start server:", err)
		return
	}
	defer tlsListener.Close()

	fmt.Println("Server active on", address+":443")
	for {
		connection, err := tlsListener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		fmt.Println("New client connected from:", connection.RemoteAddr())
		go manageClient(connection)
	}
}

func manageClient(conn net.Conn) {
	defer conn.Close()
	for {
		messageBuffer := make([]byte, 1024)
		var fullData []byte
		for {
			readBytes, err := conn.Read(messageBuffer)
			if err != nil {
				fmt.Println("Client disconnected:", conn.RemoteAddr())
				return
			}
			fullData = append(fullData, messageBuffer[:readBytes]...)
			if readBytes < len(messageBuffer) {
				break
			}
		}
		fmt.Printf("Message received (%d bytes): %s\n", len(fullData), fullData)
		response := fmt.Sprintf("Data size received: %d bytes", len(fullData))
		conn.Write([]byte(response))
	}
}

func InitTcpClient(serverAddr, privateKey, certificate string) {
	logFile, err := os.OpenFile("sslkey.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println("Failed to open key log file:", err)
		return
	}
	defer logFile.Close()

	certPair, err := tls.LoadX509KeyPair(certificate, privateKey)
	if err != nil {
		fmt.Println("Error loading TLS credentials:", err)
		return
	}

	clientConfig := tls.Config{Certificates: []tls.Certificate{certPair}, InsecureSkipVerify: true, KeyLogWriter: logFile}
	conn, err := tls.Dial("tcp", serverAddr+":8443", &clientConfig)
	if err != nil {
		fmt.Println("Connection failed:", err)
		return
	}
	defer conn.Close()

	consoleInput := bufio.NewReader(os.Stdin)
	fmt.Println("Connected to", serverAddr+":8443")

	for {
		fmt.Print("Input message to server: ")
		userInput, _ := consoleInput.ReadString('\n')
		trimmedInput := strings.TrimSpace(userInput)

		if trimmedInput == "big" {
			largeData := strings.Repeat("A", 40000)
			conn.Write([]byte(largeData))
			fmt.Println("Large payload sent.")
		} else {
			conn.Write([]byte(trimmedInput))
		}

		responseBuffer := make([]byte, 1024)
		readBytes, err := conn.Read(responseBuffer)
		if err != nil {
			fmt.Println("Error during read operation:", err)
			return
		}
		fmt.Printf("Server response: %s\n", string(responseBuffer[:readBytes]))
	}
}

func main() {
	ip := flag.String("ip", "127.0.0.1", "IP address")
	mode := flag.String("mode", "tcp-server", "Mode: tcp-server, tcp-client")
	key := flag.String("key", "", "Key file")
	crt := flag.String("crt", "", "Certificate file")
	flag.Parse()

	os.Setenv("SSLKEYLOGFILE", "sslkey.log")

	switch *mode {
	case "tcp-server":
		InitTcpServer(*ip, *key, *crt)
	case "tcp-client":
		InitTcpClient(*ip, *key, *crt)
	default:
		fmt.Println("Invalid mode provided")
	}
}
