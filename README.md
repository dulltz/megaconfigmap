# megaconfigmap
ConfigMap with no size limit.

## Description
As you may already know, ConfigMap has 1MB size limit.
However we often create configuration files larger than 1MB.

megaconfigmap enables you to manage ConfigMap larger than 1MB.

## How to use
1. Create a megaconfigmap by `kubectl megaconfigmap create <name> --from-file=<your-large-file>`
1. Create a pod with a special init-container which shares the volume of megaconfigmap.

## Quick start
1. Install kubectl-megaconfigmap in your machine.
1. Prepare a large file. `$ dd if=/dev/zero of=examples/2MB.dummy count=2 bs=1m`
1. Create a megaconfigmap. `$ kubectl megaconfigmap create <name> --from-file examples/2MB.dummy`
1. Apply example pod. `$ kubectl apply -f examples/pod.yaml`
1. Login to the pod, then you can see the 2MB.dummy. `$ kubectl exec -it megaconfigmap-demo -- ls -lh /demo/2MB.dummy`

## How it works
When you execute `kubectl megaconfigmap create`,
two types of configmap resources are generated.

- *megaconfigmap*
    - The main resource to manage partial-configmaps
    - It is mounted on init-container
    - Having the following labels:
        - `megaconfigmap.io/id`: hash string of the config file
- *partial-configmaps*
    - The children of the megaconfigmap. If you delete megaconfigmap, its children are also deleted.
    - These configmaps contain the actual content of the file specified at `--from-file`.
    - The file content is split into multiple configmaps to hold large file.
    - Having the following labels:
        - `megaconfigmap.io/id`: hash string of the config file
        - `megaconfigmap.io/order`: the ordering number of the configmap

If you can combined partial-configmaps to one, you would get the actual contents of the config file.
The init-container named *combiner* can use for it.

## Caution
Do not create too large megaconfigmap because Etcd can store only 2-3GB.
