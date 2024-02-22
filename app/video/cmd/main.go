package main

import (
	"Tiktok/app/gateway/wrapper"
	"Tiktok/app/video/dao"
	"Tiktok/app/video/script"
	"Tiktok/app/video/service"
	"Tiktok/config"
	"Tiktok/idl/video/videoPb"
	"Tiktok/mq"
	"context"
	"fmt"
	"os"

	"github.com/go-micro/plugins/v4/registry/etcd"
	ratelimit "github.com/go-micro/plugins/v4/wrapper/ratelimiter/uber"
	"github.com/go-micro/plugins/v4/wrapper/select/roundrobin"
	"github.com/go-micro/plugins/v4/wrapper/trace/opentracing"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
)

func main() {
	config.Init()
	dao.InitMySQL()
	dao.InitRedis()
	mq.InitRabbitMQ()
	loadingScript()

	defer dao.RedisClient.Close()

	// etcd注册件
	etcdReg := etcd.NewRegistry(
		registry.Addrs(fmt.Sprintf("%s:%s", config.EtcdHost, config.EtcdPort)),
	)

	// 链路追踪
	tracer, closer, err := wrapper.InitJaeger("VideoService", fmt.Sprintf("%s:%s", config.JaegerHost, config.JaegerPort))
	if err != nil {
		fmt.Printf("new tracer err: %+v\n", err)
		os.Exit(-1)
	}
	defer closer.Close()
	// 得到一个微服务实例
	microService := micro.NewService(
		micro.Name("VideoService"), // 微服务名字
		micro.Address(config.VideoServiceAddress),
		micro.Registry(etcdReg),                                  // etcd注册件
		micro.WrapHandler(ratelimit.NewHandlerWrapper(50000)),    //限流处理
		micro.WrapClient(roundrobin.NewClientWrapper()),          // 负载均衡
		micro.WrapHandler(opentracing.NewHandlerWrapper(tracer)), // 链路追踪
		micro.WrapClient(opentracing.NewClientWrapper(tracer)),   // 链路追踪
	)

	// 结构命令行参数，初始化
	microService.Init()
	// 服务注册
	_ = videoPb.RegisterVideoServiceHandler(microService.Server(), service.GetVideoSrv())
	// 启动微服务
	_ = microService.Run()
}

func loadingScript() {
	ctx := context.Background()
	go script.VideoCreateSync(ctx)
	go script.Video2RedisSync(ctx)
}
