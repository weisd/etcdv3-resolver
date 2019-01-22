# etcdv3-resolver

## server

```golang
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
```

## client

use target : etcdv3://servicename/<etcd1:2379,etcd2:2379,etcd3:2379>

```golang
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
```