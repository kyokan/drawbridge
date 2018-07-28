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

### Initiate Swap

- type: 18
- data:
	- [`2: receiving chain ID`]
	- [`32: receiving chain amount (in chain base units)`]
	- [`2: sending chain ID`]
	- [`n: sending chain commitment tx`]