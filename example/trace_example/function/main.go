package main

import (
	"context"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/transport/zipkin"
	"io"
	"os"
)

func Init(service string) (opentracing.Tracer, io.Closer) {
	transport, err := zipkin.NewHTTPTransport("http://47.103.9.218:9411/api/v1/spans", zipkin.HTTPBatchSize(1), zipkin.HTTPLogger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init zipkin: %v\n", err))
	}
	fmt.Printf("transport:%v\n", transport)

	cfg := config.Configuration{
		Sampler:  &config.SamplerConfig{Type: "const", Param: 1},
		Reporter: &config.ReporterConfig{LogSpans: true},
	}

	r := jaeger.NewRemoteReporter(transport)
	fmt.Printf("r=%v\n", r)
	tracer, closer, err := cfg.New(service,
		config.Logger(jaeger.StdLogger),
		config.Reporter(r))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func main() {
	if len(os.Args) != 2 {
		panic("ERROR: Expecting one argument")
	}
	tracer, closer := Init("function-demo")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	helloTo := os.Args[1]

	sapn := tracer.StartSpan("say-hello")
	sapn.SetTag("hello-to", helloTo)
	defer sapn.Finish()

	ctx := opentracing.ContextWithSpan(context.TODO(), sapn)

	helloStr := formString(ctx, helloTo)
	printHello(ctx, helloStr)
}

func formString(ctx context.Context, helloTo string) string {
	span, _ := opentracing.StartSpanFromContext(ctx, "formatString")
	defer span.Finish()

	helloStr := fmt.Sprintf("hello, %s!", helloTo)
	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", helloStr),
	)

	return helloStr
}

func printHello(ctx context.Context, helloStr string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "printHello")
	defer span.Finish()

	println(helloStr)
	span.LogKV("event", "println")
}
