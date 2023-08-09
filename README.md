Kluster

shenzhu.dev\
v1alpha1

generate
1. deep copy objects
2. clientset
3. informers
4. listers

```
with release 1.28

/code-generator/generate-groups.sh "deepcopy,client,lister,informer" github.com/shenzhu/kluster/pkg/client github.com/shenzhu/kluster/pkg/apis shenzhu.dev:v1alpha1 --go-header-file ./hack/boilerplate.go.txt

controller-gen paths=github.com/shenzhu/kluster/pkg/apis/shenzhu.dev/v1alpha1 crd:crdVersions=v1 crd:trivialVersions=true output:crd:artifacts:config=manifests

k create secret generic dosecret --from-literal token=[token]
```

```
apiVersion
kind
metadata:
spec:
    ...
    ...
status:
    clusterId
    progress
    kubeconfig
```