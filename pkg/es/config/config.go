// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
	elastic "gopkg.in/olivere/elastic.v5"

	"github.com/jaegertracing/jaeger/pkg/es"
	storageMetrics "github.com/jaegertracing/jaeger/storage/spanstore/metrics"
)

// Configuration describes the configuration properties needed to connect to an ElasticSearch cluster
type Configuration struct {
	Servers           []string
	Username          string
	Password          string
	Sniffer           bool          // https://github.com/olivere/elastic/wiki/Sniffing
	MaxSpanAge        time.Duration `yaml:"max_span_age"` // configures the maximum lookback on span reads
	NumShards         int64         `yaml:"shards"`
	NumReplicas       int64         `yaml:"replicas"`
	Timeout           time.Duration `validate:"min=500"`
	BulkSize          int
	BulkWorkers       int
	BulkActions       int
	BulkFlushInterval time.Duration
	IndexPrefix       string
	TagsFilePath      string
	AllTagsAsFields   bool
	TagDotReplacement string
	TLS               TLS
}

// TLS Config
type TLS struct {
	Enabled  bool
	CertPath string
	KeyPath  string
	CaPath   string
}

// ClientBuilder creates new es.Client
type ClientBuilder interface {
	NewClient(logger *zap.Logger, metricsFactory metrics.Factory) (es.Client, error)
	GetNumShards() int64
	GetNumReplicas() int64
	GetMaxSpanAge() time.Duration
	GetIndexPrefix() string
	GetTagsFilePath() string
	GetAllTagsAsFields() bool
	GetTagDotReplacement() string
}

// NewClient creates a new ElasticSearch client
func (c *Configuration) NewClient(logger *zap.Logger, metricsFactory metrics.Factory) (es.Client, error) {
	if len(c.Servers) < 1 {
		return nil, errors.New("No servers specified")
	}
	rawClient, err := elastic.NewClient(c.GetConfigs(logger)...)
	if err != nil {
		return nil, err
	}

	sm := storageMetrics.NewWriteMetrics(metricsFactory, "bulk_index")
	m := sync.Map{}

	service, err := rawClient.BulkProcessor().
		Before(func(id int64, requests []elastic.BulkableRequest) {
			m.Store(id, time.Now())
		}).
		After(func(id int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
			start, ok := m.Load(id)
			if !ok {
				return
			}
			m.Delete(id)

			// log individual errors, note that err might be false and these errors still present
			if response.Errors {
				for _, it := range response.Items {
					for key, val := range it {
						if val.Error != nil {
							logger.Error("Elasticsearch part of bulk request failed", zap.String("map-key", key),
								zap.Reflect("response", val))
						}
					}
				}
			}

			sm.Emit(err, time.Since(start.(time.Time)))
			if err != nil {
				failed := len(response.Failed())
				total := len(requests)
				logger.Error("Elasticsearch could not process bulk request",
					zap.Int("request_count", total),
					zap.Int("failed_count", failed),
					zap.Error(err),
					zap.Any("response", response))
			}
		}).
		BulkSize(c.BulkSize).
		Workers(c.BulkWorkers).
		BulkActions(c.BulkActions).
		FlushInterval(c.BulkFlushInterval).
		Do(context.Background())
	if err != nil {
		return nil, err
	}
	return es.WrapESClient(rawClient, service), nil
}

// ApplyDefaults copies settings from source unless its own value is non-zero.
func (c *Configuration) ApplyDefaults(source *Configuration) {
	if c.Username == "" {
		c.Username = source.Username
	}
	if c.Password == "" {
		c.Password = source.Password
	}
	if !c.Sniffer {
		c.Sniffer = source.Sniffer
	}
	if c.MaxSpanAge == 0 {
		c.MaxSpanAge = source.MaxSpanAge
	}
	if c.NumShards == 0 {
		c.NumShards = source.NumShards
	}
	if c.NumReplicas == 0 {
		c.NumReplicas = source.NumReplicas
	}
	if c.BulkSize == 0 {
		c.BulkSize = source.BulkSize
	}
	if c.BulkWorkers == 0 {
		c.BulkWorkers = source.BulkWorkers
	}
	if c.BulkActions == 0 {
		c.BulkActions = source.BulkActions
	}
	if c.BulkFlushInterval == 0 {
		c.BulkFlushInterval = source.BulkFlushInterval
	}
}

// GetNumShards returns number of shards from Configuration
func (c *Configuration) GetNumShards() int64 {
	return c.NumShards
}

// GetNumReplicas returns number of replicas from Configuration
func (c *Configuration) GetNumReplicas() int64 {
	return c.NumReplicas
}

// GetMaxSpanAge returns max span age from Configuration
func (c *Configuration) GetMaxSpanAge() time.Duration {
	return c.MaxSpanAge
}

// GetIndexPrefix returns index prefix
func (c *Configuration) GetIndexPrefix() string {
	return c.IndexPrefix
}

// GetTagsFilePath returns a path to file containing tag keys
func (c *Configuration) GetTagsFilePath() string {
	return c.TagsFilePath
}

// GetAllTagsAsFields returns true if all tags should be stored as object fields
func (c *Configuration) GetAllTagsAsFields() bool {
	return c.AllTagsAsFields
}

// GetTagDotReplacement returns character is used to replace dots in tag keys, when
// the tag is stored as object field.
func (c *Configuration) GetTagDotReplacement() string {
	return c.TagDotReplacement
}

// GetConfigs wraps the configs to feed to the ElasticSearch client init
func (c *Configuration) GetConfigs(logger *zap.Logger) []elastic.ClientOptionFunc {

	if c.TLS.Enabled {
		tlsConfig, err := c.CreateTLSConfig()
		if err != nil {
			return nil
		}
		httpClient := &http.Client{
			Timeout: c.Timeout,
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}

		resp, err := httpClient.Get("https://elasticsearch:9200")
		if err != nil {
			fmt.Println(err)
		}

		htmlData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		fmt.Printf("%v\n", resp.Status)
		fmt.Printf(string(htmlData))

		options := make([]elastic.ClientOptionFunc, 4)
		options[0] = elastic.SetHttpClient(httpClient)
		options[1] = elastic.SetURL(c.Servers...)
		options[2] = elastic.SetSniff(c.Sniffer)
		options[3] = elastic.SetScheme("https")
		logger.Info("tlsConfig", zap.Any("tlsConfig", tlsConfig))
		return options
	}
	options := make([]elastic.ClientOptionFunc, 3)
	options[0] = elastic.SetURL(c.Servers...)
	options[1] = elastic.SetBasicAuth(c.Username, c.Password)
	options[2] = elastic.SetSniff(c.Sniffer)
	return options
}

// CreateTLSConfig creates TLS Configuration to connect with ES Cluster.
func (c *Configuration) CreateTLSConfig() (*tls.Config, error) {
	rootCerts, err := c.LoadCertificatesFrom()
	if err != nil {
		log.Fatalf("Couldn't load root certificate from %s. Got %s.", c.TLS.CaPath, err)
	}
	if len(c.TLS.CertPath) > 0 && len(c.TLS.KeyPath) > 0 {
		clientPrivateKey, err := c.LoadPrivateKeyFrom()
		if err != nil {
			log.Fatalf("Couldn't setup client authentication. Got %s.", err)
		}
		return &tls.Config{
			RootCAs:      rootCerts,
			Certificates: []tls.Certificate{*clientPrivateKey},
		}, err
	}
	return nil, err
}

// LoadCertificatesFrom is used to load root certification
func (c *Configuration) LoadCertificatesFrom() (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(c.TLS.CaPath)
	if err != nil {
		return nil, err
	}
	certificates := x509.NewCertPool()
	certificates.AppendCertsFromPEM(caCert)
	return certificates, nil
}

// LoadPrivateKeyFrom is used to load the private certificate and key for TLS
func (c *Configuration) LoadPrivateKeyFrom() (*tls.Certificate, error) {
	privateKey, err := tls.LoadX509KeyPair(c.TLS.CertPath, c.TLS.KeyPath)
	if err != nil {
		return nil, err
	}
	return &privateKey, nil
}
