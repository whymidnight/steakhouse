# Flow


## Elements

* Funding Account
|
* Authority Account

* Staker Accounts
    - Non Fungible Token (NFT)
    - Fungible Token (Staking Coin)
|
* Escrow Account


## Testing
### Init

1) Initialize Staking Campaign with Program.

### Epoch

1) Create 2 System Accounts - Self and Holder.
2) Mint 1 NFT to Self.
3) Transfer NFT from Self to Holder.
4) Transfer _n_ Staking Coin from Self to Holder.
5) Transfer _n_ Staking Coin and 1 NFT to Staking Campaign.
420) Distribute tokens back to their respective owners.



# Cache & Scheduling

Instead of resetting `Chrono` scheduler everytime the cache is reloaded from disk - upon reception of an event, we can evaluate for `isScheduled: true` to ignore (re)scheduling the event.

This assures that such event is **only** ever scheduled once. 


# Tracing

Would there be a race condition upon **setting** `IsScheduled: true` for >1 receptions of events?

Hyptothetically, in periods of high reception frequency -

epoch == 0
------
A | 1
B | 2
C | 3

epoch == 1
------
A | 4
C | 5

epoch == 2
------
B | 6
C | 7


If we suppose:

1) `epoch == 1` and `epoch == 2` have a time displacement of _0 < displacement < 1_ seconds
2) `epoch == 1` has **not** completed a cache reload _but_ is _started_.
3) `epoch == 2` has _started_ a cache reload.

Expectation:

Since `epoch == 1` cache reload is ahead in the race to (re)load cache, it would be expected that any future cache reloads that start while prior cache (re)loads are to finish in order of sequencing.
