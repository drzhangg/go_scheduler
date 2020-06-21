package main

import (
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
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}

	r := jaeger.NewRemoteReporter(transport)
	fmt.Printf("r = %v\n", r)
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger), config.Reporter(r))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func main() {
	if len(os.Args) != 2 {
		panic("ERROR: Expecting one argument")
	}

	tracer, closer := Init("hello-world-koala")
	defer closer.Close()

	helloTo := os.Args[1]

	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", helloTo)

	helloStr := fmt.Sprintf("Hello, %s!", helloTo)
	span.LogFields(
		log.String("event", "string-format"),
		log.String("value", helloStr),
	)

	println(helloStr)
	span.LogKV("event", "println")

	span.Finish()
}
