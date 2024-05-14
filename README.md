# mlflow-operator

## Description

> This project aims to provide a Kubernetes Operator for [MLflow](https://mlflow.org/).

### Running on the local

> Check [LOCAL.md](LOCAL.md) steps for running on your local machine.

### Running on the cluster

```sh
# Install the CRDs into the cluster
make install
```

```sh
# Install instances of CR
kubectl apply -f config/samples/
```

```sh
# Build and push your image
make docker-build docker-push IMG=<some-registry>/mlflow-operator:tag
```

```sh
# Deploy the controller to the cluster
make deploy IMG=<some-registry>/mlflow-operator:tag
```

### Clean from cluster

```sh
# To delete the CRDs from the cluster
make uninstall
```

```sh
# Undeploy the controller from the cluster
make undeploy
```

### Modifying the API definitions

If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
make generate
```