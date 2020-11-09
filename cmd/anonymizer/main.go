// Copyright (c) 2020 The Jaeger Authors.
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

package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	app "github.com/jaegertracing/jaeger/cmd/anonymizer/app"
	query "github.com/jaegertracing/jaeger/cmd/anonymizer/app/query"
	writer "github.com/jaegertracing/jaeger/cmd/anonymizer/app/writer"
	"github.com/jaegertracing/jaeger/pkg/version"
)

var logger, _ = zap.NewDevelopment()

func main() {
	var options = app.Options{}

	var command = &cobra.Command{
		Use:   "jaeger-anonymizer",
		Short: "Jaeger anonymizer hashes fields of a trace for easy sharing",
		Long:  `Jaeger anonymizer queries Jaeger query for a trace, anonymizes fields, and store in file`,
		Run: func(cmd *cobra.Command, args []string) {
			prefix := options.OutputDir + "/" + options.TraceID
			conf := writer.Config{
				MaxSpansCount:    options.MaxSpansCount,
				CapturedFile:     prefix + ".original",
				AnonymizedFile:   prefix + ".anonymized",
				MappingFile:      prefix + ".mapping",
				HashStandardTags: options.HashStandardTags,
				HashCustomTags:   options.HashCustomTags,
				HashLogs:         options.HashLogs,
				HashProcess:      options.HashProcess,
			}

			writer, err := writer.New(conf, logger)
			if err != nil {
				logger.Error("error while creating writer object", zap.Error(err))
			}

			queryEndpoint := options.QueryGRPCHost + ":" + strconv.Itoa(options.QueryGRPCPort)
			logger.Info(queryEndpoint)

			query, err := query.New(queryEndpoint, logger)
			if err != nil {
				logger.Error("error while creating query object", zap.Error(err))
			}

			spans := query.QueryTrace(options.TraceID)
			for _, span := range spans {
				writer.WriteSpan(&span)
			}
		},
	}

	options.AddFlags(command)

	command.AddCommand(version.Command())

	if error := command.Execute(); error != nil {
		fmt.Println(error.Error())
		os.Exit(1)
	}
}
