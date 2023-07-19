#!/usr/bin/env bash

helm upgrade --create-namespace --install elasticsearch elastic/elasticsearch -f ./helm/values.yaml -n analytics