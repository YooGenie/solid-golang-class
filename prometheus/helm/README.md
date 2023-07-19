https://artifacthub.io/packages/helm/prometheus-community/kube-prometheus-stack

helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --create-namespace --atomic --install prometheus prometheus-community/kube-prometheus-stack -n monitoring


## Remove
### Custom Resources Removal 
for n in $(kubectl get namespaces -o jsonpath={..metadata.name}); do
  kubectl delete --all --namespace=$n prometheus,servicemonitor,podmonitor,alertmanager
done

### Custom Resources Removal Manually
kubectl delete crd alertmanagerconfigs.monitoring.coreos.com
kubectl delete crd alertmanagers.monitoring.coreos.com
kubectl delete crd podmonitors.monitoring.coreos.com
kubectl delete crd probes.monitoring.coreos.com
kubectl delete crd prometheuses.monitoring.coreos.com
kubectl delete crd prometheusrules.monitoring.coreos.com
kubectl delete crd servicemonitors.monitoring.coreos.com
kubectl delete crd thanosrulers.monitoring.coreos.com


#### Custom Resources Removal Manually
https://helm.sh/docs/chart_best_practices/custom_resource_definitions/
