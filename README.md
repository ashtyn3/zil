# Zil

_A modern version controller._

## Directory structure

```
.zil
|_  objects
    |_ hash[2]/hash[3:]0(32 bits=treeHash)(rest = base compressed diff bytes of file)
    |_ hash[2]/hash[3:]1(32 bits=treeHash)(rest = change diff bytes compressed)
|_  tree
    |_ hash(first bits = tree name)( bits = || )(rest = files sep by comma)
|_  config
    |_ text(toml file with config)
```

Objects are files repersented in diffs. Trees are just lists of hashes or just folders.

```
    zil + ./
```
