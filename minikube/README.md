# Getting Started
공식문서 참조 - https://minikube.sigs.k8s.io/docs/start/

--- 
## VirtualBox 설치하기
```
$ brew install --cask virtualbox
```

docker desktop은 이전 과정에서 설치한 것으로 간주합니다.

## Homebrew 로 설치하기
### 환경
* macOS
* x86-64 
* virtualbox 또는 docker 드라이버  
```
$ brew install minikube
```

## Minikube 클러스터 시작하기
`virtualbox` 드라이버로 클러스터 구동하기
```
$ minikube start --driver=virtualbox --cpus=4 --memory=8192
```

`docker` 드라이버로 클러스터 구동하기
```
$ minikube start --driver=docker --cpus=4 --memory=8192
```

특정 드라이버를 기본값으로 설정
```
$ minikube config set driver virtualbox
$ minikube config set driver docker
```

## Minikube 클러스터와 연결 테스트
`kubectl`이 이미 설치 되어있다면 Pod 목록 가져오기 명령어 실행
```
$ kubectl get po -A
NAMESPACE     NAME                               READY   STATUS    RESTARTS   AGE
kube-system   coredns-74ff55c5b-zz8p2            1/1     Running   1          64d
kube-system   etcd-minikube                      0/1     Running   1          64d
kube-system   kube-apiserver-minikube            1/1     Running   1          64d
kube-system   kube-controller-manager-minikube   0/1     Running   1          64d
kube-system   kube-proxy-qzwvd                   1/1     Running   1          64d
kube-system   kube-scheduler-minikube            0/1     Running   1          64d
kube-system   storage-provisioner                1/1     Running   3          64d
```
kubectl 이 없는 경우 minikube kubectl 사용
```
$ minikube kubectl -- get po -A
```
또는 kubectl 별도 설치
```
$ brew install kubectl
```

minikube kubectl 계속 사용시 alias 를 shell config에 추가
```
alias kubectl="minikube kubectl --"
```


Minikube 대시보드 접속해보기
```
$ minikube dashboard
```

## 어플리케이션 배포
샘플 `deployment(디플로이먼트)`를 생성하고 8080 포트로 외부연결 허용하기.

```
$ kubectl create deployment hello-minikube --image=k8s.gcr.io/echoserver:1.4
$ kubectl expose deployment hello-minikube --type=NodePort --port=8080
```
`service(서비스)` 상태 확인하기.
```
$ kubectl get services hello-minikube
```

브라우저로 `hello-minikube` 서비스에 접속하기
```
$ minikube service hello-minikube
```

kubtctl `port-foward` 명령어로 포트포워딩하여 접속하기

```
$ kubectl port-forward service/hello-minikube 7080:8080
```
http://localhost:7080/

