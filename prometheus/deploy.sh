#!/usr/bin/env bash


helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --create-namespace --atomic --install prometheus prometheus-community/kube-prometheus-stack -n monitoring