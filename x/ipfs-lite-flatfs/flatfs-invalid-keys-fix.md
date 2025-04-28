<img src="https://r2cdn.perplexity.ai/pplx-full-logo-primary-dark%402x.png" class="logo" width="120"/>

## How to Fix the Invalid Key Problem with FlatFS in Your Code

The **"invalid key"** error with FlatFS occurs because FlatFS only allows keys that match the pattern `/[0-9A-Z+-_=]+`-that is, only uppercase alphanumeric characters and `-`, `+`, `_`, `=` are allowed. **Namespaced keys** (those with slashes, like `/blocks/XYZ`) are *not* allowed[^3][^5]. If any part of your code or its dependencies tries to store such a key, FlatFS will panic or return an error.

### Why This Happens in Your Code

- **ipfs-lite** is generally safe, as it uses CIDs (which are conformant) as keys for block storage.
- However, if you use the *same FlatFS datastore* for other purposes (e.g., as a generic datastore for metadata, state, or anything that uses namespaced keys), you will hit this error.
- Some higher-level APIs or libraries may try to store keys like `/blocks/XYZ` or `/foo/bar`, which FlatFS will reject.

---

## Solution: Use FlatFS *Only* for Block Storage

**Do not use FlatFS as your general-purpose datastore.**

- Use FlatFS only as the blockstore (where CIDs are used as keys).
- For anything else (metadata, state, etc.), use a different datastore backend, such as Badger, LevelDB, or in-memory datastores, which support namespaced keys.


### How to Fix Your Code

#### 1. Use FlatFS Only for the Blockstore

- **ipfs-lite.New** accepts separate datastores for block storage and for metadata/state.
- Pass FlatFS only as the blockstore, and use a different, compatible datastore (e.g., Badger or LevelDB) for the rest.


#### 2. Example: Split Datastores

```go
import (
    // ... other imports
    badgerds "github.com/ipfs/go-ds-badger2"
    ds "github.com/ipfs/go-datastore"
)

// ... inside main()

// FlatFS for blockstore
fs, err := flatfs.CreateOrOpen(flatfsDir, shard, false)
if err != nil {
    panic(err)
}
blockstoreDS := ds_sync.MutexWrap(fs)

// Badger for metadata/state
badgerDir := "/tmp/ipfs-lite-badger"
os.MkdirAll(badgerDir, 0755)
metaDS, err := badgerds.NewDatastore(badgerDir, &amp;badgerds.DefaultOptions)
if err != nil {
    panic(err)
}

// Pass both to ipfs-lite
peer, err := ipfslite.New(ctx, blockstoreDS, metaDS, host, kaddht, nil)
if err != nil {
    panic(err)
}
```

- Here, `blockstoreDS` is your FlatFS instance (for blocks only), and `metaDS` is a Badger datastore for everything else.
- This ensures that only conformant keys go to FlatFS, and any namespaced keys (used for metadata, pinning, etc.) go to Badger.

---

## Summary Table

| Datastore Purpose | Backend to Use | Key Format Allowed |
| :-- | :-- | :-- |
| Block Storage (blocks) | FlatFS | CIDs (conformant) |
| Metadata/State | Badger, LevelDB, etc. | Namespaced keys supported |


---

## Final Notes

- **Never use FlatFS as a general-purpose datastore**-only as a blockstore[^3][^5].
- If you must use a single datastore for everything, pick one that supports namespaced keys (not FlatFS).
- This split-datastore approach is standard in IPFS and is how go-ipfs and ipfs-lite are designed to work.

---

**Implementing this change will fix the invalid key problem and prevent FlatFS panics or errors in your application.**

<div style="text-align: center">⁂</div>

[^1]: https://ppl-ai-file-upload.s3.amazonaws.com/web/direct-files/attachments/40300957/b6e1d46e-7149-4033-aa87-92fc1b1a04cd/ipfs-lite-flatfs.go

[^2]: https://github.com/ipfs/go-ipfs/issues/7149

[^3]: https://github.com/ipfs/go-ds-flatfs

[^4]: https://discuss.ipfs.tech/t/ipfs-key-list-error/15858

[^5]: https://pkg.go.dev/github.com/ipfs/go-ds-flatfs

[^6]: https://discuss.ipfs.tech/t/problems-with-access-control-mechanisms/16522

[^7]: https://github.com/ipfs/notes/issues/206

[^8]: https://stackoverflow.com/questions/75641616/how-can-i-fix-invalid-key-error-retrieve-json-data-of-specific-places-from-m

[^9]: https://discuss.ipfs.tech/t/what-is-in-the-ipfs-folders-datastore/9204

[^10]: https://stackoverflow.com/questions/54270908/how-to-decode-ipfs-private-and-public-key-in-der-pem-format

[^11]: https://discuss.ipfs.tech/t/fixed-unknown-datastore-type-flatfs/15805

[^12]: https://www.reddit.com/r/ipfs/comments/s9fr9m/storing_private_files_on_ipfs_that_can_only_be/

[^13]: https://github.com/hsanjuan/ipfs-lite

[^14]: https://github.com/ipfs/kubo/issues/8993

[^15]: https://discuss.ipfs.tech/t/unixfs-object-overhead/7606

[^16]: https://docs.ipfs.eth.link/how-to/configure-node/

[^17]: https://stackoverflow.com/questions/78188969/err-encryption-failed-with-libp2p-connecting-to-kubo-ipfs

[^18]: https://forum.sia.tech/t/grant-proposal-ipfs-sia-renterd-ipfsr/404

[^19]: https://stackoverflow.com/questions/77726166/nextauth-and-ipfs-error-when-hosting-on-app-fleek-co

[^20]: https://ipfs.github.io/js-stores/modules/interface_datastore.html

[^21]: https://discuss.ipfs.tech/t/importing-pem-encoded-private-key/12770

