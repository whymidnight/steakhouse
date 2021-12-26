# Goals


### Showcase Event Consumption

1) Initiate Smart Wallet w/
    - Timed reward function transaction of 60 seconds
2) Upsert NFTs into Smart Wallet from Test Wallet(s)
3) Await airdrop of rewards upon expiry

### Status

1223 - update

1) Need to implement transactionWithTimelock instead of wallet with `MinimumDelay`
2) Using devnet hunters - still need to design the following:

    - An integration test conducting:
    * Create smart wallet and derived wallet address
    > Create smart wallet **without** `miniumDelay`
    > Create Transaction **with+* Timelock
    > Defer Transaction Execution over RPC

    * Upsert non-fungible tokens to derived wallet address
    > Specify a wallet with n mints selecting y mints
    > Upsert selected mints into _derived_ smart wallet address
