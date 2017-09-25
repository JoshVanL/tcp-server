package main

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
	"k8s.io/client-go/1.5/kubernetes"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/rest"
)

const nameSpace = "namespace-a"
const addr = ":8800"
const kind = "tcp"
const service = "pod-server-svc"

func main() {
	// Setup listener
	listener, err := net.Listen(kind, addr)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Listening on " + addr)

	for {
		// Acc connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
		}

		fmt.Printf("Connection opened: %v\n", conn.LocalAddr().String())

		// Handle connection -- concurrent
		go handleReq(conn)
	}
}

func handleReq(conn net.Conn) {
	// Read buffer from conn
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Skipping request... error reading message: %v", err)

	} else {
		// Get pod ane svc using input as pod name
		resp := string(buf[:n])
		podName := strings.TrimSpace(resp)
		pod, svc, err := get(podName)
		respond(conn, podName, pod, svc, err)
	}

	if err := conn.Close(); err != nil {
		fmt.Printf("Error closing connection: %v", err)
	} else {
		fmt.Printf("Connection closed: %v\n", conn.LocalAddr().String())
	}
}

func get(podName string) (podP *v1.Pod, svcP *v1.Service, err error) {
	var errs error

	// Read this Kubernetes config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	// Get pod using input as pod name
	pod, err := c.Pods(nameSpace).Get(podName)
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Error getting pod by name: %v", err))
	}

	// Get service
	svc, err := c.Services(nameSpace).Get(service)
	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("Error getting service by name: %v", err))
	}

	return pod, svc, errs
}

// Send back response
func respond(conn net.Conn, podName string, pod *v1.Pod, svc *v1.Service, err error) {
	conn.Write([]byte(fmt.Sprintf("Getting names from pod: %s\n", podName)))

	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Errors: %s\n", err)))

	} else {
		if pod != nil {
			conn.Write([]byte(fmt.Sprintf("Names: %s, %s, %s\n", pod.GetName(), pod.GetNamespace(), pod.GetCreationTimestamp().String())))
		}

		if svc != nil {
			conn.Write([]byte(fmt.Sprintf("Service: %v\n", svc.GetName())))
		}
	}
}
