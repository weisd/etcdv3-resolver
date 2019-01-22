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

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"

	_ "github.com/weisd/etcdv3_resolver"
)

var (
	etcdAddrs   string
	defaultName string
)

func init() {
	flag.StringVar(&etcdAddrs, "etcd_addrs", "127.0.0.1:2379", "etcd endpoints eg: 127.0.0.1:2379")
	flag.StringVar(&defaultName, "name", "world", "say hello args")
}

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(
		fmt.Sprintf("etcdv3://test/%s", etcdAddrs),
		grpc.WithBalancerName("round_robin"),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Contact the server and print out its response.
	name := defaultName

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c := pb.NewGreeterClient(conn)
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("%s-%d", name, i)

		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s\n", r.Message)
	}

}
