# deployment-deletor

## Prerequisites

https://github.com/kubernetes-sigs/kubebuilder/blob/book-v3/docs/book/src/quick-start.md#prerequisites

## Install CRD

```
make install
```

## Deploy

```
make deploy
```

## Delete

```
make undeploy
```

## Uninstall CRD

`make undeploy` includes CRD uninstallation

```
make uninstall
```
