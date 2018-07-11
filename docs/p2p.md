# Peer-to-Peer Messaging

Drawbridge manages an instance of an `lnd` node. Below, see the description of the wire format for each message.

Every message has a global prefix of `0xbeef`. This is designed to prevent collisions with normal Lightning messages in the future.

After the prefix, every message contains the following payload:

- type: [`2: big-endian number`]
- data: [`n: variable length field`]

The following messages are supported:

## Initialization Message

- type: 16
- data:
	- [`33: local lnd identification key`]
	- [`2: lnd addr length`]
	- [`lnd addr length: lnd addr`]

### Requirements

The receiving node:

- MUST respond with another `init` message before sending any other messages.
- MUST wait for the local `lnd` node to connect to the remote `lnd` node, otherwise fail the connection.

## Control Messages

`ping` and `pong` messages behave identically to how they behave in `lnd`.

## Swap Messages

### Initiate Swap (ERC-20/ETH for BTC)

- type: 900
- data:
 	- [`32: swap ID`]
 	- [`32: payment hash`]
	- [`32: ETH channel ID`]
	- [`32: ETH amount`]
	- [`65: ETH commitment signature`]
	- [`20: sending address`]
	- [`32: receiving amount`]

### Swap Accepted

- type: 901
- data:
	- [`32: swap ID`]
	- [`8: BTC channel ID`]

### Invoice Generated

- type: 902
- data:
	- [`32: swap ID`]
	- [`8: invoice len`]
	- [`invoice len: invoice`]

### Invoice Executed

Executes invoice upon receipt.

- type: 903
- data:
	- 	[`32: swap ID`]