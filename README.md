# Reconfigure-StorageClass-Annotation

Service to set the annotation that marks a StorageClass as default to false.
A StorageClass is set to default when the annotation "storageclass.kubernetes.io/is-default-class" is set to "true".

## Description

If you want to create a persistent volume on a [Bottlerocket](https://github.com/bottlerocket-os/bottlerocket) host, you will need to use the [EBS CSI Plugin](https://github.com/kubernetes-sigs/aws-ebs-csi-driver). This is because the default EBS driver relies on file system tools that are not included with Bottlerocket. This Service allows you to automatically set the StorageClass "gp2" to not default.

With this service you can create a Terraform module that deploys it on Kubernetes. This allows you do automatically set the StorageClass to not default, rather then doing it manually.

## Usage

Below is an example YAML file to create all necessary resources.

```YML
apiVersion: v1
kind: ServiceAccount
metadata:
  name: reconfigure-storageclass-annotation
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: reconfigure-storageclass-annotation
rules:
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses"]
  verbs: ["list", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: reconfigure-storageclass-annotation
subjects:
- kind: ServiceAccount
  name: reconfigure-storageclass-annotation
  apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: ClusterRole
  name: reconfigure-storageclass-annotation
  apiGroup: rbac.authorization.k8s.io
---
apiVersion: v1
kind: Pod
metadata:
  name: reconfigure-storageclass-annotation
spec:
  serviceAccountName: reconfigure-storageclass-annotation
  containers:
  - name: reconfigure-storageclass-annotation
    image: docker.pkg.github.com/timkra/reconfigure-storageclass-annotation/reconfigure-storageclass-annotation
    env:
    - name: STORAGE_CLASS_NAME
      value: "gp2"
```

This service requires permissions to "list" and "patch" a storage class.
Using RBAC authorization, the permission can be granted.

The example creates a ServiceAccount, and binds a ClusterRole to the ServiceAccount. The ServiceAccount is reference on the pod spec.

The pod itself only pulls the container image. Using the environment variable "STORAGE_CLASS_NAME" we specify the StorageClass to be reconfigured.
