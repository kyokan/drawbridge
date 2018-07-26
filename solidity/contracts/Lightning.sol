pragma solidity 0.4.24;

import "./ERC20.sol";
import "./BytesLib.sol";
import "./BytesBuffer.sol";
import "zeppelin-solidity/contracts/ECRecovery.sol";

contract Lightning {
    using BytesLib for bytes;
    using BytesBuffer for BytesBuffer.Buffer;
    
    bytes1 PAYMENT_SIGIL = 0x01;
    
    bytes1 MULTISIG_SIGIL = 0x02;
    
    bytes1 LOCAL_COMMIT_SIGIL = 0x03;
    
    bytes1 HTLC_RECV_SIGIL = 0x05;
    
    bytes1 HTLC_OFFER_SIGIL = 0x06;
    
    bytes ZERO_SIG = new bytes(65);
    
    bytes ZERO_BYTES = new bytes(0);
    
    address ZERO_ADDRESS = address(0);
    
    bytes32 ZERO_BYTES32 = bytes32(0);
    
    uint ENCUMBRANCE_SIZE = 290;
    
    address public tokenAddress;
    
    uint public lastTxId;
    
    struct Output {
        uint lockTime;
        uint value;
        uint blockNum;
        bytes script;
        uint id;
        bool exists;
    }
    
    mapping(uint => Output) public outputs;
    
    event Withdrawal(address owner, uint value);
    
    event NewOutput(uint lockTime, uint value, uint blockNum, bytes script, uint id);
    
    event DebugBytes(bytes b);
    //
    event DebugBytes32(bytes32 b);
    //
    event DebugUint(uint b);
    //
    event DebugAddress(address a);
    
    event DebugBytes1(bytes1 a);
    
    constructor(address _tokenAddress) public {
        tokenAddress = _tokenAddress;
        lastTxId = 0;
    }
    
    function deposit(uint value) public {
        require(value > 0);
        
        ERC20 tokenContract = ERC20(tokenAddress);
        require(tokenContract.transferFrom(msg.sender, address(this), value));
        
        bytes memory script = paymentScript(msg.sender);
        
        Output memory out = Output(
            0,
            value,
            block.number,
            script,
            ++lastTxId,
            true
        );
        
        outputs[out.id] = out;
        emitNewOutput(out);
    }
    
    function withdraw(bytes witness, address payer) public {
        require(witness.length == 133);
        uint totalInputValue = processInputs(witness, ZERO_BYTES);
        ERC20 tokenContract = ERC20(tokenAddress);
        require(tokenContract.transfer(payer, totalInputValue));
        emit Withdrawal(payer, totalInputValue);
    }
    
    function spend(bytes witnesses, bytes outputScripts) public {
        uint totalInputValue = processInputs(witnesses, outputScripts);
        
        // lockTime (32), value (32), sigil(1), script (n)
        
        uint cursor = 0;
        uint totalOutputValue = 0;
        
        while (cursor < outputScripts.length) {
            uint lockTime = outputScripts.toUint(cursor);
            cursor += 32;
            uint value = outputScripts.toUint(cursor);
            cursor += 32;
            bytes1 sigil = outputScripts[cursor];
            uint16 scriptLen = scriptLength(sigil);
            require(scriptLen > 0);
            require(outputScripts.length - cursor >= scriptLen);
            bytes memory script = outputScripts.slice(cursor, scriptLen);
            cursor += scriptLen;
            
            totalOutputValue += value;
            
            Output memory out = Output(
                lockTime,
                value,
                block.number,
                script,
                ++lastTxId,
                true
            );
            
            outputs[out.id] = out;
            emitNewOutput(out);
        }
        
        require(totalInputValue == totalOutputValue);
    }
    
    function processInputs(bytes witnesses, bytes outputScripts) private returns (uint) {
        uint cursor = 0;
        uint totalInputValue = 0;
        
        // outputid (32), witnesslen (16), witness, sig
        while (cursor < witnesses.length) {
            uint startScriptSig = 0;
            uint outputId = witnesses.toUint(cursor);
            cursor += 32;
            uint16 witnessLen = witnesses.toUint16(cursor);
            cursor += 16;
            bytes memory witness = witnesses.slice(cursor, witnessLen);
            cursor += witnessLen;
            uint endScriptSig = cursor;
            bytes memory sig = witnesses.slice(cursor, 65);
            bytes memory scriptSig = witnesses.slice(startScriptSig, cursor - startScriptSig);
            Output memory input = outputs[outputId];
            require(isSpendable(input, scriptSig, outputScripts, sig));
            totalInputValue += input.value;
            cursor += 65;
            delete outputs[input.id];
        }
        
        return totalInputValue;
    }
    
    function outputId(uint value, bytes script, bytes32 inputId) private returns (bytes32) {
        return keccak256(block.number, value, script, inputId);
    }
    
    function verify(bytes32 data, address expected, bytes sig) returns (bool) {
        return ECRecovery.recover(ECRecovery.toEthSignedMessageHash(data), sig) == expected;
    }
    
    function paymentScript(address redeemer) public returns (bytes) {
        BytesBuffer.Buffer memory buf = BytesBuffer.Buffer(new bytes(21), 0);
        buf.putByte(PAYMENT_SIGIL);
        buf.putAddress(redeemer);
        return buf.data;
    }
    
    function emitNewOutput(Output out) private {
        emit NewOutput(out.lockTime, out.value, out.blockNum, out.script, out.id);
    }
    
    function isSpendable(Output out, bytes witnessScript, bytes outputScripts, bytes sig) private returns (bool) {
        bytes1 sigil = out.script[0];
        bytes32 checkHash = keccak256(witnessScript, outputScripts);
        
        if (sigil == PAYMENT_SIGIL) {
            return executePayment(out, witnessScript, checkHash, sig);
        }
        
        return false;
    }
    
    function executePayment(Output output, bytes witnessScript, bytes32 checkHash, bytes sig) private returns (bool) {
        address redeemer = output.script.toAddress(1);
        return redeemer == witnessScript.toAddress(witnessScript.length - 20) && verify(checkHash, redeemer, sig);
    }
    
    function scriptLength(bytes1 sigil) public returns (uint16) {
        if (sigil == PAYMENT_SIGIL) {
            return 21;
        }
        
        return 0;
    }
}