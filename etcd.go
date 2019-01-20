package etcd

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
	"google.golang.org/grpc/resolver"
)

var _ resolver.Builder = &Builder{}
var _ resolver.Resolver = &Resolver{}

// Scheme grpc 调用的target scheme
const Scheme = "etcdv3"

// Prefix etcd kv 存储的key前缀
const Prefix = "/grpc_etcdv3_resolver"

var (
	// DefaultTimeout etcd
	DefaultTimeout = time.Second * 5
)

func init() {
	resolver.Register(&Builder{})
}

// Builder Builder
type Builder struct {
}

// Build implement resolver.Builder  启动程序监控target数据变化
// scheme://authority/endpoint eg. etcdv3://grpcserver/localhost:2379,localhost:22379,localhost:32379
func (p *Builder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	// 从target中拿到etcd的服务器地址

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(target.Endpoint, ","),
		DialTimeout: DefaultTimeout,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := &Resolver{
		cc:     cc,
		cli:    cli,
		ctx:    ctx,
		cancel: cancel,
		key:    filepath.Join(Prefix, target.Authority, "/"),
	}

	go r.watch()

	return r, nil
}

// Scheme implement resolver.Builder
func (p *Builder) Scheme() string {
	return Scheme
}

// Resolver Resolver
type Resolver struct {
	cc  resolver.ClientConn
	cli *clientv3.Client

	ctx    context.Context
	cancel context.CancelFunc

	key string
}

// ResolveNow implement resolver.Resolver
func (p *Resolver) ResolveNow(o resolver.ResolveNowOption) {}

// Close implement resolver.Resolver
func (p *Resolver) Close() {}

// 服务器地址存储的key
func (p *Resolver) watchKey() string {
	return p.key
}

// watch 监控etcd数据变化
func (p *Resolver) watch() {
	rch := p.cli.Watch(p.ctx, p.watchKey())
	for {
		select {
		case <-p.ctx.Done():
			return
		case wresp := <-rch:
			if wresp.IsProgressNotify() {
				continue
			}

			// 重新get更新
			p.syncAddress()
		}

	}

}

// 通知
func (p *Resolver) syncAddress() {
	ctx, cancel := context.WithTimeout(p.ctx, DefaultTimeout)

	resp, err := p.cli.Get(ctx, p.watchKey())
	cancel()
	if err != nil {
		fmt.Println("etcd client get err", err)
		return
	}

	addrs := make([]resolver.Address, len(resp.Kvs))
	for i, ev := range resp.Kvs {
		fmt.Printf("%s : %s\n", ev.Key, ev.Value)
		addrs[i] = resolver.Address{Addr: string(ev.Key)}

	}
	p.cc.NewAddress(addrs)
}

// Register Register
type Register struct {
	cli *clientv3.Client

	ctx    context.Context
	cancel context.CancelFunc

	key  string
	addr string
}

// New New
func (p *Register) New(name, addr string, endpoints []string) (*Register, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: DefaultTimeout,
	})
	if err != nil {
		return nil, err
	}

	return &Register{
		cli:  cli,
		key:  filepath.Join(Prefix, name, "/"),
		addr: addr,
	}, nil
}

// Key 服务器地址存储的key
func (p *Register) Key() string {
	return filepath.Join(p.key, p.addr)
}

// Registe 注册
func (p *Register) Registe(ctx context.Context) error {
	_, err := p.cli.Put(ctx, p.Key(), "")
	if err != nil {
		return err
	}

	return nil
}

// DelRegiste 删除
func (p *Register) DelRegiste(ctx context.Context, addr string) error {
	_, err := p.cli.Delete(ctx, p.Key())
	if err != nil {
		return err
	}

	return nil
}

// AutoRegisteWithExpire 自动注册，过期自动删除
// sec 有效时间 单位：秒
func (p *Register) AutoRegisteWithExpire(ctx context.Context, sec int64) error {

	resp, err := p.cli.Grant(ctx, sec)
	if err != nil {
		return err
	}

	_, err = p.cli.Put(ctx, p.Key(), "", clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	// the key 'foo' will be kept forever
	ch, kaerr := p.cli.KeepAlive(ctx, resp.ID)
	if kaerr != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
			case ka := <-ch:
				fmt.Println("ttl:", ka.TTL)
			}
		}
	}()

	return nil
}

// Close Close
func (p *Register) Close() {
	if p.cli != nil {
		p.cli.Close()
	}
}
