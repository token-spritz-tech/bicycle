## TODO
- [x] Withdraw TON method
- [x] Withdraw jetton method
- [x] Generate new address API method
- [x] Get addresses for user API method
- [x] Send TON/Jetton API method
- [x] Shard block scanner and tx parsing
- [x] Batched send TONs method
- [x] Batched send Jettons method
- [x] TON transfer comment saving
- [x] TON deposit withdrawal
- [x] Custom withdrawal comment support
- [x] Jetton transfer comment saving
- [x] Jetton deposit withdrawal
- [x] Deposit withdrawal validation
- [x] Restart policy (repair after reconnect)
- [x] Time sync with node
- [x] Cold wallets support
- [x] Shard merge/split detecting
- [x] Graceful shutdown
- [x] Healthcheck
- [x] License
- [x] Deploy (in need) hot wallet on start 
- [x] Queue/webhook notifications
- [x] Deposits balances get method
- [x] Deposit balance calculation flag (after deposit filling or hot-wallet filling)
- [x] Threat model draft
- [x] Anomalous behavior detecting and audit log
- [x] Docs
- [x] Refactoring
- [x] Deploy scripts
- [x] Unit tests
- [x] Validate wallets code for tonutils v1.4.1
- [x] Manual testing plan
- [x] Service methods for API (cancellation of incorrect payments)
- [x] Build emulator lib from sources
- [x] Integration tests
- [x] Hot wallets metrics
- [x] Manual testing
- [x] Jettons test list
- [x] Fix timeouts
- [x] Allow to start with empty Jetton env var
- [x] Deposit side balances by default
- [x] Fix "outgoing message from internal incoming" for bounced TON payment 
- [x] Add history method
- [x] Rename balance to income and return owner address instead of jetton wallet (for queue too)
- [x] Add history method to test plan
- [x] Add filling deposit with bounce to test plan
- [x] Update to tonutils-go 1.6.2
- [x] Process masterchain addresses for external incomes
- [x] Cold wallet withdrawal fix
- [x] Add hysteresis to cold wallet withdrawal
- [x] Add user id to notifications
- [x] Add transaction hash to notifications
- [ ] Avoid blocking withdrawals to an address if there is a very large amount in the queue for withdrawals to this address
- [ ] Save tx hash to DB
- [ ] Support DNS names in recipient address
- [ ] Jetton threat model
- [ ] TNX compatibility test
- [ ] Installation video manual
- [ ] Use stable branch for emulator
- [ ] Download blockchain config at start
- [ ] Add reconnect to node when timeout expires
- [ ] Node deploy
- [ ] Performance optimization
- [ ] Fix base64 public key format in .env file
- [ ] Describe recovery scenarios
- [ ] BOLT compatibility test
- [ ] Not process removed Jettons
- [ ] Separate .env files for services
- [ ] Automatic migrations
- [ ] SDK
- [ ] migration from blueprint to openapi
- [ ] refactor config and cutoff parameters