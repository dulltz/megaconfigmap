# megaconfigmap

As you may already know, ConfigMap has 1MB size limit.  
However we often create configuration files larger than 1MB.👼

**megaconfigmap** enables you to manage ConfigMap larger than 1MB.

This system consists of two components:
- `combiner` - One-shot program to combine partial comfigmaps to large one. It is designed to run on init-container. 
- `kubectl-megaconfigmap` - kubectl plugin to create large congfigmap.

## Quick start
1. Install kubectl-megaconfigmap.
   ```console
   $ git clone git@github.com:dulltz/megaconfigmap.git
   $ cd megaconfigmap
   $ go install -mod=vendor ./cmd/kubectl-megaconfigmap/
   ```
1. Prepare a large file.
   ```console
   $ dd if=/dev/zero of=examples/2MB.dummy count=2 bs=1m
   ```
1. Create a megaconfigmap.
   ```console
   $ kubectl megaconfigmap create my-conf --from-file examples/2MB.dummy
   ```
1. Apply the example manifest.
   ```console
   $ kubectl apply -f examples/pod.yaml
   ```
1. Login to the pod, then you can see the 2MB.dummy. 
   ```console
   $ kubectl exec -it megaconfigmap-demo -- ls -lh /demo
   ```

## How to use
1. Create a megaconfigmap by `kubectl megaconfigmap create <name> --from-file=<your-large-file>`
1. Create a pod with a special init-container which shares the volume of megaconfigmap.

## How it works
1. Create megaconfigmap and partial-configmaps by `kubectl megaconfigmap create`.
1. Combiner init-container collect partial-item from megaconfigmap specified at `--megaconfigmap` flag.
1. Combiner dump the file to the path on the share volume specified at `--share-dir` flag.
1. If you mount the share volume to the main container, you can get the large file there. 

## Glossary
- *megaconfigmap*
    - The owner of partial-configmaps
    - It is not mounted
    - It has only metadata
    - Having the following labels:
        - `megaconfigmap.io/id`: hash string of the config file
        - `megaconfigmap.io/filename`: output file name
        - `megaconfigmap.io/master`: indicate that this resource is a megaconfigmap
- *partial-configmaps*
    - The children of the megaconfigmap. If you delete megaconfigmap, its children are also deleted.
    - These configmaps contain the partial data of source file.
    - The file content is split into multiple configmaps to hold large file.
    - Having the following labels:
        - `megaconfigmap.io/id`: hash string of the config file
        - `megaconfigmap.io/filename`: output file name
        - `megaconfigmap.io/order`: the ordering number of the configmap

## Caution
Do not create too large megaconfigmap because Etcd can store only 2-3GB.
