package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	// User's provided printer IP
	printerIP := "192.168.123.10"
	printerPort := "9100" 
	address := net.JoinHostPort(printerIP, printerPort)

	fmt.Println("=================================================")
	fmt.Println("       KAFERUN PRINTER DIAGNOSTICS TOOL        ")
	fmt.Println("=================================================")
	fmt.Printf("DATE:    %s\n", time.Now().Format("02.01.2006 15:04:05"))
	fmt.Printf("TARGET:  %s\n", address)
	fmt.Println("-------------------------------------------------")

	// Step 1: TCP Handshake
	fmt.Printf("[1/2] Attempting to connect to printer... ")
	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	
	if err != nil {
		fmt.Println("FAILED ❌")
		fmt.Printf("\nERROR DETAILS:\n%v\n", err)
		
		fmt.Println("\nPOSSIBLE SOLUTIONS:")
		fmt.Println("1. Check if the printer is powered ON.")
		fmt.Println("2. Verify that this computer and the printer are on the SAME network/WiFi.")
		fmt.Printf("3. Double-check if the printer's IP is actually %s.\n", printerIP)
		fmt.Println("4. Ensure port 9100 is not blocked by a local firewall.")
		fmt.Println("=================================================")
		return
	}
	defer conn.Close()
	fmt.Println("SUCCESS ✅")

	// Step 2: Data Transfer
	fmt.Printf("[2/2] Sending ESC/POS test command... ")
	
	// ESC/POS Commands
	init := []byte{0x1B, 0x40}      // Init
	cut := []byte{0x1D, 0x56, 0x42, 0x00} // Paper Cut
	feed := []byte("\n\n\n\n\n")
	content := []byte("\n================================\n")
	content = append(content, []byte("        NETWORK TEST OK!       \n")...)
	content = append(content, []byte("      Server: 46.224.133.140   \n")...)
	content = append(content, []byte("================================\n")...)

	payload := append(init, content...)
	payload = append(payload, feed...)
	payload = append(payload, cut...)

	_, err = conn.Write(payload)
	if err != nil {
		fmt.Println("FAILED ❌")
		fmt.Printf("Data error: %v\n", err)
	} else {
		fmt.Println("SUCCESS ✅")
		fmt.Println("\nRESULT: Communication with the printer is fully functional!")
	}
	fmt.Println("=================================================")
}
