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
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"
	"gopkg.in/olivere/elastic.v5"

	"github.com/jaegertracing/jaeger/pkg/es"
	"github.com/jaegertracing/jaeger/pkg/es/wrapper"
	storageMetrics "github.com/jaegertracing/jaeger/storage/spanstore/metrics"
)

// Configuration describes the configuration properties needed to connect to an ElasticSearch cluster
type Configuration struct {
	Servers             []string
	Username            string
	Password            string
	TokenFilePath       string
	Sniffer             bool          // https://github.com/olivere/elastic/wiki/Sniffing
	MaxNumSpans         int           // defines maximum number of spans to fetch from storage per query
	MaxSpanAge          time.Duration `yaml:"max_span_age"` // configures the maximum lookback on span reads
	NumShards           int64         `yaml:"shards"`
	NumReplicas         int64         `yaml:"replicas"`
	Timeout             time.Duration `validate:"min=500"`
	BulkSize            int
	BulkWorkers         int
	BulkActions         int
	BulkFlushInterval   time.Duration
	IndexPrefix         string
	TagsFilePath        string
	AllTagsAsFields     bool
	TagDotReplacement   string
	Enabled             bool
	TLS                 TLSConfig
	UseReadWriteAliases bool
}

// TLSConfig describes the configuration properties to connect tls enabled ElasticSearch cluster
type TLSConfig struct {
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
	GetMaxNumSpans() int
	GetIndexPrefix() string
	GetTagsFilePath() string
	GetAllTagsAsFields() bool
	GetTagDotReplacement() string
	GetUseReadWriteAliases() bool
	GetTokenFilePath() string
	IsEnabled() bool
}

// NewClient creates a new ElasticSearch client
func (c *Configuration) NewClient(logger *zap.Logger, metricsFactory metrics.Factory) (es.Client, error) {
	if len(c.Servers) < 1 {
		return nil, errors.New("No servers specified")
	}
	options, err := c.getConfigOptions()
	if err != nil {
		return nil, err
	}

	rawClient, err := elastic.NewClient(options...)
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
			if response != nil && response.Errors {
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
				var failed int
				var respval interface{}
				if response == nil {
					failed = 0
					respval = "nil"
				} else {
					failed = len(response.Failed())
					respval = response
				}
				total := len(requests)
				logger.Error("Elasticsearch could not process bulk request",
					zap.Int("request_count", total),
					zap.Int("failed_count", failed),
					zap.Error(err),
					zap.Any("response", respval))
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
	return eswrapper.WrapESClient(rawClient, service), nil
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
	if c.MaxNumSpans == 0 {
		c.MaxNumSpans = source.MaxNumSpans
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

// GetMaxNumSpans returns max spans allowed per query from Configuration
func (c *Configuration) GetMaxNumSpans() int {
	return c.MaxNumSpans
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

// GetUseReadWriteAliases indicates whether read alias should be used
func (c *Configuration) GetUseReadWriteAliases() bool {
	return c.UseReadWriteAliases
}

// GetTokenFilePath returns file path containing the bearer token
func (c *Configuration) GetTokenFilePath() string {
	return c.TokenFilePath
}

// IsEnabled determines whether storage is enabled
func (c *Configuration) IsEnabled() bool {
	return c.Enabled
}

// getConfigOptions wraps the configs to feed to the ElasticSearch client init
func (c *Configuration) getConfigOptions() ([]elastic.ClientOptionFunc, error) {
	options := []elastic.ClientOptionFunc{elastic.SetURL(c.Servers...), elastic.SetSniff(c.Sniffer)}
	httpClient := &http.Client{
		Timeout: c.Timeout,
	}
	options = append(options, elastic.SetHttpClient(httpClient))
	if c.TLS.Enabled {
		ctlsConfig, err := c.TLS.createTLSConfig()
		if err != nil {
			return nil, err
		}
		httpClient.Transport = &http.Transport{
			TLSClientConfig: ctlsConfig,
		}
	} else {
		httpTransport := &http.Transport{}
		if c.TLS.CaPath != "" {
			ctls := &TLSConfig{CaPath: c.TLS.CaPath}
			ca, err := ctls.loadCertificate()
			if err != nil {
				return nil, err
			}
			httpTransport.TLSClientConfig = &tls.Config{RootCAs: ca}
		}
		if c.TokenFilePath != "" {
			token, err := loadToken(c.TokenFilePath)
			if err != nil {
				return nil, err
			}
			httpClient.Transport = &tokenAuthTransport{
				token:   token,
				wrapped: httpTransport,
			}
		} else {
			httpClient.Transport = httpTransport
			options = append(options, elastic.SetBasicAuth(c.Username, c.Password))
		}
	}
	return options, nil
}

// createTLSConfig creates TLS Configuration to connect with ES Cluster.
func (tlsConfig *TLSConfig) createTLSConfig() (*tls.Config, error) {
	rootCerts, err := tlsConfig.loadCertificate()
	if err != nil {
		return nil, err
	}
	clientPrivateKey, err := tlsConfig.loadPrivateKey()
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs:      rootCerts,
		Certificates: []tls.Certificate{*clientPrivateKey},
	}, nil

}

// loadCertificate is used to load root certification
func (tlsConfig *TLSConfig) loadCertificate() (*x509.CertPool, error) {
	caCert, err := ioutil.ReadFile(tlsConfig.CaPath)
	if err != nil {
		return nil, err
	}
	certificates := x509.NewCertPool()
	certificates.AppendCertsFromPEM(caCert)
	return certificates, nil
}

// loadPrivateKey is used to load the private certificate and key for TLS
func (tlsConfig *TLSConfig) loadPrivateKey() (*tls.Certificate, error) {
	privateKey, err := tls.LoadX509KeyPair(tlsConfig.CertPath, tlsConfig.KeyPath)
	if err != nil {
		return nil, err
	}
	return &privateKey, nil
}

// TokenAuthTransport
type tokenAuthTransport struct {
	token   string
	wrapped *http.Transport
}

func (tr *tokenAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Authorization", "Bearer "+tr.token)
	return tr.wrapped.RoundTrip(r)
}

func loadToken(path string) (string, error) {
	b, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(b), "\r\n"), nil
}
