// Copyright (c) 2019 The Jaeger Authors.
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
	"errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/olivere/elastic"
	"github.com/uber/jaeger-lib/metrics"
	"go.uber.org/zap"

	"github.com/jaegertracing/jaeger/pkg/config/tlscfg"
	"github.com/jaegertracing/jaeger/pkg/es"
	eswrapper "github.com/jaegertracing/jaeger/pkg/es/wrapper"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	storageMetrics "github.com/jaegertracing/jaeger/storage/spanstore/metrics"
)

type RolloverInterval byte

const (
	// RolloverDaily means creating an new index once a day, this is the default behavior
	RolloverDaily = iota
	// RolloverHourly means creating an new index once an hour
	RolloverHourly
	// RolloverQuarterly means creating an new index once a quarter of an hour
	RolloverQuarterly
)

// Configuration describes the configuration properties needed to connect to an ElasticSearch cluster
type Configuration struct {
	Servers               []string      `mapstructure:"server_urls"`
	Username              string        `mapstructure:"username"`
	Password              string        `mapstructure:"password"`
	TokenFilePath         string        `mapstructure:"token_file"`
	AllowTokenFromContext bool          `mapstructure:"-"`
	Sniffer               bool          `mapstructure:"sniffer"`               // https://github.com/olivere/elastic/wiki/Sniffing
	MaxNumSpans           int           `mapstructure:"-"`                     // defines maximum number of spans to fetch from storage per query
	MaxSpanAge            time.Duration `yaml:"max_span_age" mapstructure:"-"` // configures the maximum lookback on span reads
	NumShards             int64         `yaml:"shards" mapstructure:"num_shards"`
	NumReplicas           int64         `yaml:"replicas" mapstructure:"num_replicas"`
	Timeout               time.Duration `validate:"min=500" mapstructure:"-"`
	BulkSize              int           `mapstructure:"-"`
	BulkWorkers           int           `mapstructure:"-"`
	BulkActions           int           `mapstructure:"-"`
	BulkFlushInterval     time.Duration `mapstructure:"-"`
	IndexPrefix           string        `mapstructure:"index_prefix"`
	RolloverInterval      RolloverInterval
	Tags                  TagsAsFields   `mapstructure:"tags_as_fields"`
	Enabled               bool           `mapstructure:"-"`
	TLS                   tlscfg.Options `mapstructure:"tls"`
	UseReadWriteAliases   bool           `mapstructure:"use_aliases"`
	CreateIndexTemplates  bool           `mapstructure:"create_mappings"`
	Version               uint           `mapstructure:"version"`
}

// TagsAsFields holds configuration for tag schema.
// By default Jaeger stores tags in an array of nested objects.
// This configurations allows to store tags as object fields for better Kibana support.
type TagsAsFields struct {
	// Store all tags as object fields, instead nested objects
	AllAsFields bool `mapstructure:"all"`
	// Dot replacement for tag keys when stored as object fields
	DotReplacement string `mapstructure:"dot_replacement"`
	// File path to tag keys which should be stored as object fields
	File string `mapstructure:"config_file"`
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
	GetRolloverInterval() RolloverInterval
	IsStorageEnabled() bool
	IsCreateIndexTemplates() bool
	GetVersion() uint
}

// NewClient creates a new ElasticSearch client
func (c *Configuration) NewClient(logger *zap.Logger, metricsFactory metrics.Factory) (es.Client, error) {
	if len(c.Servers) < 1 {
		return nil, errors.New("no servers specified")
	}
	options, err := c.getConfigOptions(logger)
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
				if response == nil {
					failed = 0
				} else {
					failed = len(response.Failed())
				}
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

	if c.Version == 0 {
		// Determine ElasticSearch Version
		pingResult, _, err := rawClient.Ping(c.Servers[0]).Do(context.Background())
		if err != nil {
			return nil, err
		}
		esVersion, err := strconv.Atoi(string(pingResult.Version.Number[0]))
		if err != nil {
			return nil, err
		}
		logger.Info("Elasticsearch detected", zap.Int("version", esVersion))
		c.Version = uint(esVersion)
	}

	return eswrapper.WrapESClient(rawClient, service, c.Version), nil
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
	if c.RolloverInterval == RolloverDaily {
		c.RolloverInterval = source.RolloverInterval
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
	return c.Tags.File
}

// GetAllTagsAsFields returns true if all tags should be stored as object fields
func (c *Configuration) GetAllTagsAsFields() bool {
	return c.Tags.AllAsFields
}

// GetVersion returns Elasticsearch version
func (c *Configuration) GetVersion() uint {
	return c.Version
}

// GetTagDotReplacement returns character is used to replace dots in tag keys, when
// the tag is stored as object field.
func (c *Configuration) GetTagDotReplacement() string {
	return c.Tags.DotReplacement
}

// GetUseReadWriteAliases indicates whether read alias should be used
func (c *Configuration) GetUseReadWriteAliases() bool {
	return c.UseReadWriteAliases
}

// GetTokenFilePath returns file path containing the bearer token
func (c *Configuration) GetTokenFilePath() string {
	return c.TokenFilePath
}

// IsStorageEnabled determines whether storage is enabled
func (c *Configuration) IsStorageEnabled() bool {
	return c.Enabled
}

// IsCreateIndexTemplates determines whether index templates should be created or not
func (c *Configuration) IsCreateIndexTemplates() bool {
	return c.CreateIndexTemplates
}

// getConfigOptions wraps the configs to feed to the ElasticSearch client init
func (c *Configuration) getConfigOptions(logger *zap.Logger) ([]elastic.ClientOptionFunc, error) {

	options := []elastic.ClientOptionFunc{elastic.SetURL(c.Servers...), elastic.SetSniff(c.Sniffer),
		// Disable health check when token from context is allowed, this is because at this time
		// we don' have a valid token to do the check ad if we don't disable the check the service that
		// uses this won't start.
		elastic.SetHealthcheck(!c.AllowTokenFromContext)}
	httpClient := &http.Client{
		Timeout: c.Timeout,
	}
	options = append(options, elastic.SetHttpClient(httpClient))
	if c.TLS.Enabled {
		ctlsConfig, err := c.TLS.Config()
		if err != nil {
			return nil, err
		}
		httpClient.Transport = &http.Transport{
			TLSClientConfig: ctlsConfig,
		}
	} else {
		httpTransport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			// #nosec G402
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.TLS.SkipHostVerify},
		}
		if c.TLS.CAPath != "" {
			config, err := c.TLS.Config()
			if err != nil {
				return nil, err
			}
			httpTransport.TLSClientConfig = config
		}

		token := ""
		if c.TokenFilePath != "" {
			if c.AllowTokenFromContext {
				logger.Warn("Token file and token propagation are both enabled, token from file won't be used")
			}
			tokenFromFile, err := loadToken(c.TokenFilePath)
			if err != nil {
				return nil, err
			}
			token = tokenFromFile
		}

		if token != "" || c.AllowTokenFromContext {
			httpClient.Transport = &tokenAuthTransport{
				token:                token,
				allowOverrideFromCtx: c.AllowTokenFromContext,
				wrapped:              httpTransport,
			}
		} else {
			httpClient.Transport = httpTransport
			options = append(options, elastic.SetBasicAuth(c.Username, c.Password))
		}
	}
	return options, nil
}

// GetRolloverInterval returns rollover interval
func (c *Configuration) GetRolloverInterval() RolloverInterval {
	return c.RolloverInterval
}

// TokenAuthTransport
type tokenAuthTransport struct {
	token                string
	allowOverrideFromCtx bool
	wrapped              *http.Transport
}

func (tr *tokenAuthTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	token := tr.token
	if tr.allowOverrideFromCtx {
		headerToken, _ := spanstore.GetBearerToken(r.Context())
		if headerToken != "" {
			token = headerToken
		}
	}
	r.Header.Set("Authorization", "Bearer "+token)
	return tr.wrapped.RoundTrip(r)
}

func loadToken(path string) (string, error) {
	b, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(b), "\r\n"), nil
}
