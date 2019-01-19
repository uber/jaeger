module github.com/jaegertracing/jaeger

require (
	github.com/PuerkitoBio/purell v0.0.0-20171117214151-1c4bec281e4b
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/Shopify/sarama v1.17.0
	github.com/VividCortex/gohistogram v1.0.0
	github.com/apache/thrift v0.0.0-20151001171628-53dd39833a08
	github.com/asaskevich/govalidator v0.0.0-20171128153514-852d82c746b2
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973
	github.com/bsm/sarama-cluster v2.1.13+incompatible
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd
	github.com/crossdock/crossdock-go v0.0.0-20160816171116-049aabb0122b
	github.com/davecgh/go-spew v1.1.1
	github.com/eapache/go-resiliency v1.1.0
	github.com/eapache/go-xerial-snappy v0.0.0-20160609142408-bb955e01b934
	github.com/eapache/queue v0.0.0-20180227141424-093482f3f8ce
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-kit/kit v0.5.0
	github.com/go-openapi/analysis v0.0.0-20171208011258-5c7230aa5ab8
	github.com/go-openapi/errors v0.0.0-20170426151106-03cfca65330d
	github.com/go-openapi/jsonpointer v0.0.0-20170102174223-779f45308c19
	github.com/go-openapi/jsonreference v0.0.0-20161105162150-36d33bfe519e
	github.com/go-openapi/loads v0.0.0-20171207192234-2a2b323bab96
	github.com/go-openapi/runtime v0.0.0-20171207053002-55d76b231921
	github.com/go-openapi/spec v0.0.0-20171206193454-01738944bdee
	github.com/go-openapi/strfmt v0.0.0-20170822153411-610b6cacdcde
	github.com/go-openapi/swag v0.0.0-20171111214437-cf0bdb963811
	github.com/go-openapi/validate v0.0.0-20171117174350-d509235108fc
	github.com/gocql/gocql v0.0.0-20181124151448-70385f88b28b
	github.com/gogo/gateway v1.0.0
	github.com/gogo/googleapis v0.0.0-20180501115203-b23578765ee5
	github.com/gogo/protobuf v0.0.0-20171130202109-fd9a4790f396
	github.com/golang/protobuf v0.0.0-20180202184318-bbd03ef6da3a
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db
	github.com/gorilla/context v1.1.1
	github.com/gorilla/handlers v0.0.0-20161206055144-3a5767ca75ec
	github.com/gorilla/mux v1.3.0
	github.com/grpc-ecosystem/grpc-gateway v0.0.0-20180312001938-58f78b988bc3
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed
	github.com/hashicorp/hcl v0.0.0-20171017181929-23c074d0eceb
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/kr/pretty v0.0.0-20160823170715-cfb55aafdaf3
	github.com/magiconair/properties v0.0.0-20171031211101-49d762b9817b
	github.com/mailru/easyjson v0.0.0-20171120080333-32fa128f234d
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/mitchellh/mapstructure v0.0.0-20171017171808-06020f85339e
	github.com/mwitkow/go-proto-validators v0.0.0-20180403085117-0950a7990007
	github.com/olivere/elastic v5.0.39+incompatible
	github.com/opentracing-contrib/go-stdlib v0.0.0-20171029140428-b1a47cfbdd75
	github.com/opentracing/opentracing-go v1.0.2
	github.com/pelletier/go-toml v0.0.0-20171024211038-4e9e0ee19b60
	github.com/pierrec/lz4 v2.0.2+incompatible
	github.com/pkg/errors v0.8.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v0.9.1
	github.com/prometheus/client_model v0.0.0-20171117100541-99fa1f4be8e5
	github.com/prometheus/common v0.0.0-20180326160409-38c53a9f4bfc
	github.com/prometheus/procfs v0.0.0-20180321230812-780932d4fbbe
	github.com/rakyll/statik v0.0.0-20180522225057-eab4de8dac7e
	github.com/rcrowley/go-metrics v0.0.0-20180503174638-e2704e165165
	github.com/spf13/afero v0.0.0-20171021110813-5660eeed305f
	github.com/spf13/cast v1.1.0
	github.com/spf13/cobra v0.0.1
	github.com/spf13/jwalterweatherman v0.0.0-20170901151539-12bd96e66386
	github.com/spf13/pflag v1.0.2
	github.com/spf13/viper v1.1.0
	github.com/stretchr/objx v0.0.0-20180702103455-b8b73a35e983
	github.com/stretchr/testify v1.2.2
	github.com/uber-go/atomic v1.3.1
	github.com/uber/jaeger-client-go v2.14.0+incompatible
	github.com/uber/jaeger-lib v1.5.0
	github.com/uber/tchannel-go v1.1.0
	go.uber.org/atomic v1.3.2
	go.uber.org/multierr v1.1.0
	go.uber.org/zap v1.9.1
	golang.org/x/net v0.0.0-20180811021610-c39426892332
	golang.org/x/sys v0.0.0-20160516132347-d4feaf1a7e61
	golang.org/x/text v0.0.0-20171102192421-88f656faf3f3
	google.golang.org/genproto v0.0.0-20171212231943-a8101f21cf98
	google.golang.org/grpc v1.11.0
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/mgo.v2 v2.0.0-20160818020120-3f83fa500528
	gopkg.in/olivere/elastic.v5 v5.0.53
	gopkg.in/yaml.v2 v2.0.0
	jaeger-storage-dynamodb v0.0.0
)

replace jaeger-storage-dynamodb => ../../ledor473/jaeger-storage-dynamodb
