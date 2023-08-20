# auth-center

// TODO(user): Add simple overview of use/purpose
创建项目: `kubebuilder init --domain jiayi.com --repo jiayi.com/auth-center`
创建api: `kubebuilder create api --group auth --version v1 --kind AuthCenter`
kubebuilder create webhook --group auth --version v1 --kind AuthCenter --defaulting --programmatic-validation

## 设计思路

> 1. 第一版版本主要考虑用户的只读和读写权限属性，主要涉及为默认两个clusterRole，一个是只读，一个是读写，然后通过RoleBinding来绑定用户和clusterRole
> 2. 第二版项目角色-用户权限关系， 权限授权给项目角色，项目项目下面的用户可以有不同的权限。
> 3. 第三版项目角色-操作Action-用户，项目权限力度细化

1. 登陆的用户跟RoleBinding 中的subject :User关联如

```yaml
subjects:
  - apiGroup: rbac.authorization.k8s.io
    kind: User
    name: vin1
```

2. AuthCenter主键为用户的id信息，然后去关联对应的k8s权限和harbor资源权限
3. 通过修改、创建AuthCenter资源，会同步去更新对应的k8sRolebinding资源以及Harbor资源
4. 回去watch 所有的RoleBinding资源，将对应的权限统一写到AuthCenter这个CRD的status资源信息中
5. 轮询harbor信息，同步到CRD资源中
6. 使用validate校验用户的权限，如果没有权限，就不允许创建对应的资源 如没有命名空间，或则没有harbor项目
7. 删除命名空间- 》 会在对应的命名空间添加一个finalizer， 然后watch 命名空间， 如果删除命名空间 则修改对应的AuthCenter资源

## Description

// TODO(user): An in-depth paragraph about your project and overview of use

## Getting Started

You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for
testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever
cluster `kubectl cluster-info` shows).

### Running on the cluster

修改了api的字段值需要进行 make manifests 重新生成对应的代码
安装crd make install

1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/auth-center:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/auth-center:tag
```

### Uninstall CRDs

To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller

UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing

// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works

This project aims to follow the
Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the
cluster.

### Test It Out

1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

