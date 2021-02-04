#!/bin/bash

# this script expects all docker images to be already built, it only uploads them to Docker Hub

set -euxf -o pipefail

BRANCH=${BRANCH:?'missing BRANCH env var'}

# Only push images to dockerhub/quay.io for master branch or for release tags vM.N.P
if [[ "$BRANCH" == "master" || $BRANCH =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "upload to dockerhub/quay.io, BRANCH=$BRANCH"
else
  echo 'skip docker images upload, only allowed for tagged releases or master (latest tag)'
  exit 0
fi

export DOCKER_NAMESPACE=jaegertracing

jaeger_components=(
	agent
	agent-debug
	cassandra-schema
	es-index-cleaner
	es-rollover
	collector
	collector-debug
	query
	query-debug
	ingester
	ingester-debug
	tracegen
	anonymizer
	opentelemetry-collector
	opentelemetry-agent
	opentelemetry-ingester
)

DOCKERHUB_USERNAME=${DOCKERHUB_USERNAME:-}
DOCKERHUB_TOKEN=${DOCKERHUB_TOKEN:-}
QUAY_USERNAME=${QUAY_USERNAME:-}
QUAY_TOKEN=${QUAY_TOKEN:-}

for component in "${jaeger_components[@]}"
do
  REPO="jaegertracing/jaeger-${component}"
  bash scripts/travis/upload-to-registry.sh "docker.io" $REPO $DOCKERHUB_USERNAME $DOCKERHUB_TOKEN
  bash scripts/travis/upload-to-registry.sh "quay.io" $REPO $QUAY_USERNAME $QUAY_TOKEN
done
