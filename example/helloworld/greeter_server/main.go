/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
	"strings"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/reflection"

	resolver "github.com/weisd/etcdv3_resolver"
)

const ()

var (
	etcdAddrs string
	port      int
)

func init() {
	flag.StringVar(&etcdAddrs, "etcd_addrs", "127.0.0.1:2379", "etcd endpoints eg: 127.0.0.1:2379")
	flag.IntVar(&port, "port", 8001, "server port")
}

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Printf("Received: %v\n", in.Name)
	return &pb.HelloReply{Message: "Hello " + in.Name}, nil
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
			cancel()
		}
	}()

	addr := fmt.Sprintf("%s:%d", localIP(), port)

	register, err := resolver.NewRegister("test", addr, strings.Split(etcdAddrs, ","))
	if err != nil {
		log.Fatalln(err)
	}

	register.AutoRegisteWithExpire(ctx, 5)

	select {
	case <-ctx.Done():
		cancel()

	}
}

func localIP() string {
	name := "eth0"
	if runtime.GOOS == "darwin" {
		name = "en0"
	}

	itf, _ := net.InterfaceByName(name)

	addrs, _ := itf.Addrs()
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}

		if strings.Index(ip.String(), "::") > -1 {
			continue
		}
		// process IP address
		return ip.String()
	}

	return "localhost"
}
