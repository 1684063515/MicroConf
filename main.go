package main

import (
	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/etcd"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-plugins/registry/etcdv3/v2"
	"log"
)

var (
	Endpoint = []string{"127.0.0.1:2379", "127.0.0.1:22379", "127.0.0.1:32379"}
)

type Config struct {
	Hosts struct {
		Database struct {
			Address string `json:"address"`
			Port    int    `json:"port"`
		} `json:"database"`
		Cache struct {
			Address string `json:"address"`
			Port    int    `json:"port"`
		} `json:"cache"`
	} `json:"hosts"`
}

func main() {
	conf, _ := InitCfg()
	GetConf(conf)

	reg := etcdv3.NewRegistry(func(opts *registry.Options) {
		// etcd 集群地址
		opts.Addrs = []string{"127.0.0.1:2379", "127.0.0.1:22379", "127.0.0.1:32379"}
	})

	// micro服务
	service := micro.NewService(
		micro.Name("greeter"),
		micro.Registry(reg),
	)

	service.Init()

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}

// 加载与监听配置修改
func InitCfg() (config.Config, error) {
	source := etcd.NewSource(
		etcd.WithAddress(Endpoint...),
		etcd.WithPrefix("/micro/config/mysql"),
	)

	conf, _ := config.NewConfig()
	if err := conf.Load(source); err != nil {
		log.Fatal("load error:", err)
	}

	go func() {
		watcher, err := conf.Watch("micro", "config", "mysql")
		if err != nil {
			log.Fatal("watch error:", err)
			return
		}

		for {
			v, err := watcher.Next()
			if err != nil {
				log.Fatal("watch next error:", err)
				return
			}
			log.Printf("value:%+v", string(v.Bytes()))
		}
	}()

	return conf, nil
}

// 获取配置
func GetConf(conf config.Config) {
	cf := &Config{}
	v := conf.Get("micro", "config", "mysql")
	if v != nil {
		v.Scan(&cf)
	} else {
		log.Printf("配置不存在")
	}
	log.Printf("host:%+v", cf)
}
