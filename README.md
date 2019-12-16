# yorc-storage-plugin
A yorc storage sample plugin where deployments are stored in the file system

This sample shows how to create a Go plugin to externalize Yorc storage for a defined store type.

Here is implemented a file system store to save all deployments data in a specified directory.

## The store plugin in few steps

- Define an exported plugin "symbol" Store

```
 var Store fileStore
 ```
 
- In the init function of main.go, instantiate the Store previously declared

- The store implementation has to implement yorc storage.Store interface (file_store.go)
Note the Types() method determines which store type(s) is concerned

```
func (s *fileStore) Types() []types.StoreType {
	t := make([]types.StoreType, 0)
	t = append(t, types.StoreTypeDeployment)
	return t
}
```


## How to build

- Build yorc locally with the CGO enabled (export CGO_ENABLED=1)
- Set the yotc sources path in the go.mod file (replace cmd)
- Build this plugin (make)
