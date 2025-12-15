# Journal

On-chain we store:
batchCommitment = hash(rootHash || totalEntries || timestampRangeHash || version)

- rootHash = Merkle root of all entries
- totalEntries = number of entries in batch
- timestampRangeHash = hash(startTimestamp || endTimestamp)
- version = protocol version

Off-Chain Key-Value Database Schema

1. Batch Registry (registry:batch:{commitment})

```json
{
  "rootHash": "0xabc123...",
  "totalEntries": 150,
  "startTime": 1633023456000,
  "endTime": 1633023556000,
  "treeDepth": 8,
  "leafFormat": "sha256(data || position)",
  "commitment": "batchCommitment (on-chain hash)",
  "status": "finalized",
  "createdAt": 1633023557000
}
```

2.Entry Storage (Content-Addressable)

```t
Key: e:{dataHash}
Value: {
  "data": "actual entry content",
  "position": 3,
  "batchCommitment": "batchCommitment",
  "timestamp": 1633023456050,
  "metadata": {}  // optional additional data
}

```

3.Position Index (Fast lookup by position)

```t
Key: p:{batchCommitment}:{position:08d}
Value: dataHash  // points to entry key
```

4.Merkle Proof Cache (Optimized for verification) (Not fully implemented)

```t
Key: proof:{batchCommitment}:{position}
Value: {
  "leafHash": "hash of entry at position",
  "siblings": ["sibling1", "sibling2", ...],
  "path": [0,1,0,...]  // 0=left,1=right
}
```

5.Reverse Lookup (Find batch by entry)

```t
Key: r:{dataHash}
Value: batchCommitment
```

Complete Key Structure

```python
# All keys used prefixes 
BATCH_PREFIX = "b:"
ENTRY_PREFIX = "e:"
POSITION_PREFIX = "p:"
PROOF_PREFIX = "proof:"
REVERSE_PREFIX = "r:"
TIME_PREFIX = "t:"

# Example keys:
batch_key = f"b:{batchCommitment}"
entry_key = f"e:{dataHash}"
position_key = f"p:{batchCommitment}:{position:08d}"
proof_key = f"proof:{batchCommitment}:{position}"
reverse_key = f"r:{dataHash}"
```
