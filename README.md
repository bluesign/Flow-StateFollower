# Flow State Follower ( pub/sub for events / txs / blocks / storage changes ) 

Note: Only events implemented so far. 

## subsctibe to event 

```json
{
  "action": "subscribe",
  "topic": "A.f919ee77447b7497.FlowFees.FeesDeducted"
}
```


## sample output:

```
{"type":"Event","value":{"id":"A.1e3c78c6d580273b.LNVCT.Withdraw","fields":[{"name":"id","value":{"type":"UInt64","value":"21542628243865601"}},{"name":"from","value":{"type":"Optional","value":{"type":"Address","value":"0x1e3c78c6d580273b"}}}]}}
```
