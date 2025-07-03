package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
	"tunneler/pkg/gcloud"
	"tunneler/pkg/tunnel"
)

func main() {
	project := flag.String("project", "", "The Google Cloud project ID")
	zone := flag.String("zone", "", "The GCP zone where the VM is located")
	instance := flag.String("instance", "", "The name of the VM to tunnel through")
	sqlInstance := flag.String("sql-instance", "", "The name of the Cloud SQL instance")
	localPort := flag.Int("local-port", 1433, "The local port to listen on")
	sqlPort := flag.Int("sql-port", 3307, "The port of the Cloud SQL instance")
	sshUser := flag.String("ssh-user", "", "The SSH user for the VM")
	sshKeyFile := flag.String("ssh-key-file", "", "The path to the SSH private key file")

	flag.Parse()

	if *project == "" || *zone == "" || *instance == "" || *sqlInstance == "" || *sshUser == "" || *sshKeyFile == "" {
		log.Fatal("All flags are required")
	}

	key, err := os.ReadFile(*sshKeyFile)
	if err != nil {
		log.Fatalf("Failed to read SSH key file: %v", err)
	}

	privateKey, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("Failed to parse SSH private key: %v", err)
	}

	sqlIP, err := gcloud.GetSQLInstanceIP(context.Background(), *project, *sqlInstance)
	if err != nil {
		log.Fatalf("Failed to get Cloud SQL instance IP: %v", err)
	}

	dialOptions := tunnel.DialOptions{
		Project:  *project,
		Zone:     *zone,
		Instance: *instance,
		Port:     22, // SSH port
	}

	log.Printf("Starting tunnel to %s:%d through %s in %s/%s", sqlIP, *sqlPort, *instance, *project, *zone)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *localPort))
	if err != nil {
		log.Fatalf("Failed to listen on local port: %v", err)
	}
	defer listener.Close()

	for {
		localConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept new connection: %v", err)
			continue
		}

		go func() {
			defer localConn.Close()

			iapConn, err := tunnel.Dial(context.Background(), nil, dialOptions)
			if err != nil {
				log.Printf("Failed to dial IAP connection: %v", err)
				return
			}
			defer iapConn.Close()

			sshTunnel, err := tunnel.NewSSHTunnel(iapConn, *sshUser, privateKey)
			if err != nil {
				log.Printf("Failed to establish SSH tunnel: %v", err)
				return
			}
			defer sshTunnel.Close()

			remoteConn, err := sshTunnel.Dial("tcp", fmt.Sprintf("%s:%d", sqlIP, *sqlPort))
			if err != nil {
				log.Printf("Failed to dial remote connection through SSH: %v", err)
				return
			}
			defer remoteConn.Close()

			go io.Copy(remoteConn, localConn)
			io.Copy(localConn, remoteConn)
		}()
	}
}
