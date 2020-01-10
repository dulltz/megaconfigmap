# mega-configmap
ConfigMap with no size limit.

## Description
As you may already know, ConfigMap has 1MB size limit.
However we often create configuration files larger than 1MB.

mega-configmap enables you to manage ConfigMap larger than 1MB.

## How to use
1. Create a mega-configmap by `kubectl megaconfigmap create <name> --from-file=<your-large-file>`
1. Create a pod with a special init-container which shares the volume of mega-configmap.

## Quick start
1. `$ kubectl megaconfigmap create <name> --from-file examples/2MB-config.json`
1. `$ kubectl apply -f examples/pod.yaml`
1. Login to the pod, then you can see the 2MB-config.json.

    ```console
    $ kubectl exec -it megaconfigmap-example -- ls -l /data/2MB-config.json
    ```

## How it works
When you execute `kubectl megaconfigmap create`,
two types of configmap resources are generated.

- *mega-configmap*
    - The main resource to manage partial-configmaps
    - It is mounted on init-container
    - Having the following labels:
        - `mega-configmap.io/id`: hash string of the config file
- *partial-configmaps*
    - The children of the mega-configmap. If you delete mega-configmap, its children are also deleted.
    - These configmaps contain the actual content of the file specified at `--from-file`.
    - The file content is split into multiple configmaps to hold large file.
    - Having the following labels:
        - `mega-configmap.io/id`: hash string of the config file
        - `mega-configmap.io/order`: the ordering number of the configmap

If you can combined partial-configmaps to one, you would get the actual contents of the config file.
The init-container named *combiner* can use for it.

## Caution
Do not create too large mega-configmap because Etcd can store only 2-3GB.
