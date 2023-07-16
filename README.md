# scalyr-k8s-operator

Work in progress. 

## Description

`scalyr-k8s-operator` is a kubernetes operator to control and reconcile Scalyr resources in a kubernetes-native way. Built around [The Operator Framework](https://operatorframework.io) - an open source toolkit to manage Kubernetes native applications, called Operators, in an effective, automated, and scalable way.

## Roadmap

  * [x] POC: management of Scalyr Alerts and Scalyr Alert configuration file
  * [ ] Scalyr log parsers
  * [ ] Scalyr dashboards
  * [ ] Anything else that would be needed or requested

## How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster.

## Contribution

Go through [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html) and this [tutorial](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/) that should cover the basics.
This is a [multigroup](https://book.kubebuilder.io/migration/multi-group.html) operator.

Quickstart:
```
operator-sdk create api --group myGroup --version v1alpha1 --kind MyResourceKind --resource --controller
make manifests
make generate
make install
...
```

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
