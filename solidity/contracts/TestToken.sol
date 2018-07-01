pragma solidity 0.4.24;

import 'zeppelin-solidity/contracts/token/ERC20/MintableToken.sol';

contract TestToken is MintableToken {
    string public name = "Test Token";

    string public symbol = "TEST";

    uint8 public decimals = 18;
}