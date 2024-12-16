# GOPay


### Advantages of this Design:
1. **Seamless Audit Trail**: 
* `Transaction` captures real-money movement, with clear links to `Payment` for context. 
2. **Extensibility**: 
* Future features like fee structures, batching, or conditions are easily supported via `Transaction Meta`
3. **Flexibility in Roles**:
* Use PaymentIdentity.RoleName to model custom relationships (e.g., splitting fees, platform intermediaries).
4. **Custom Workflow**:
*Supports manual and programmatic payment workflows through linked transactions.