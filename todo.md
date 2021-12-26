# TODO


## Delete processed events from disk? At least retain processed events on disk

* It is unknown if fsnotify is buggy but what we can do -

1) Generate UUID at time of event subscription (client-side)
2) Include UUID as idemptotent key in request to RPC endpoint
3) Include `{isScheduled,isProcessed}` fields to RPC request struct

### Then,

1) Mutate `{isScheduled,isProcessed}` in data struct with `true`
2) Filter for `isProcessed: false && isScheduled: false` upon _scheduling_ of events
    > The reason why we filter `isScheduled: false` is because goroutines can not be cancelled without a channel.
    > To counteract this, we employ the following:

    * Upon spawing the goroutine to consume the event, the corresponding file is mutated with `isScheduled: true`.
    * Upon completion of the contained process in said goroutine, the corresponding file is mutated (again) with `isProcessed: true`.

## Finally,

The CLI argument `--recover` _should_ **reset** file(s) that were `isScheduled: true && isProcessed: false` to `isScheduled: false && isProcessed: false` before `events.loadCache` as it would be out of scope to pass down the explicit definition of the `--recovery`.

## Takeaway

[untested] * `--recover` needs to invoke a blocking call to reset stale cached files.
[halfway] * Implement `Chrono` scheduler ???


## Updates

Need to incorporate an instruction to `transfer` n NFTs to _derived_ wallet address.

Need to correspond `TransactionCreateEvent` - ie **Reward Tx** - to some business logic that:

* `GetTokenAccountsByOwner` of the Derived Smart Wallet Address.
* Create _n_ instructions to represent _n_ staked NFTs that need dismissal back to original owners.



### Cost structure

#### Minimum Cost: >_n_ Tokens


getStakes = getTokenAccountsByOwner(1)
getStake = getSignaturesForAddress(1) + getTransaction(1)




