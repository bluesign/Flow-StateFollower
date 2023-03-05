### Flow State Follower ( pub/sub for events / txs / blocks / storage changes ) 

Note: Only events implemented so far. 

use https://websocketking.com to test.

#### subscribe to event 

```json
{
  "action": "subscribe",
  "topic": "A.1e3c78c6d580273b.LNVCT.Withdraw"
}
```


#### sample output:

```
{"type":"Event","value":{"id":"A.1e3c78c6d580273b.LNVCT.Withdraw","fields":[{"name":"id","value":{"type":"UInt64","value":"21542628243865601"}},{"name":"from","value":{"type":"Optional","value":{"type":"Address","value":"0x1e3c78c6d580273b"}}}]}}
```
