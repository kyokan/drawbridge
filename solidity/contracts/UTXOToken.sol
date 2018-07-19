pragma solidity 0.4.24;

import "./ERC20.sol";
import "./BytesLib.sol";
import "./BytesBuffer.sol";
import "zeppelin-solidity/contracts/ECRecovery.sol";

contract UTXOToken {
    using BytesLib for bytes;
    using BytesBuffer for BytesBuffer.Buffer;

    bytes1 MULTISIG_SIGIL = 0x02;

    bytes ZERO_SIG = new bytes(65);

    bytes ZERO_BYTES = new bytes(0);

    address ZERO_ADDRESS = address(0);

    bytes32 ZERO_BYTES32 = bytes32(0);

    uint ENCUMBRANCE_SIZE = 290;

    address public tokenAddress;

    struct UTXO {
        address owner;
        uint value;
        uint blockNum;
        bytes32 id;
        bool exists;
    }

    struct Multisig {
        address signerA;
        address signerB;
        uint valueA;
        uint valueB;
        uint blockNum;
        bytes32 id;
        bytes32 hashLock;
        uint lockTime;
        address lockedFor;
        bool exists;
    }

    struct Encumbrance {
        bytes32 id;
        uint lockTime;
        uint valueA;
        uint valueB;
        bytes32 hashLock;
        bytes sigA;
        bytes sigB;
    }

    mapping(bytes32 => UTXO) public utxos;

    mapping(bytes32 => Multisig) public multisigs;

    event Deposit(address owner, uint value, bytes32 id);

    event Withdrawal(address owner, uint value, bytes32 id);

    event Output(bytes32 inputId, address owner, uint value, bytes32 id);

    event MultisigOutput(address signerA, address signerB, uint valueA, uint valueB, bytes32 id);

    event Breach(address breachingAddress, bytes32 multisigId);

//    event DebugBytes(bytes b);
//
//    event DebugBytes32(bytes32 b);
//
//    event DebugUint(uint b);
//
//    event DebugAddress(address a);

    constructor(address _tokenAddress) public {
        tokenAddress = _tokenAddress;
    }

    function deposit(uint value) public {
        require(value > 0);

        address owner = msg.sender;
        ERC20 tokenContract = ERC20(tokenAddress);
        require(tokenContract.transferFrom(owner, address(this), value));

        UTXO memory utxo = makeUTXO(owner, value, bytes32(0));
        utxos[utxo.id] = utxo;
        emit Deposit(owner, value, utxo.id);
        emit Output(ZERO_BYTES32, owner, value, utxo.id);
    }

    function withdraw(bytes32 id) public {
        UTXO memory input = utxos[id];
        address owner = msg.sender;
        require(input.exists);
        require(input.owner == owner);

        ERC20 tokenContract = ERC20(tokenAddress);
        require(tokenContract.transfer(owner, input.value));

        delete utxos[input.id];
        emit Withdrawal(owner, input.value, input.id);
    }

    function spend(bytes32 id, address recipient, uint value) public {
        require(value > 0);

        UTXO memory input = utxos[id];
        require(input.exists);
        require(input.owner == msg.sender);
        require(input.value >= value);

        UTXO memory spent = makeUTXO(recipient, value, input.id);
        utxos[spent.id] = spent;

        if (value < input.value) {
            UTXO memory change = makeUTXO(msg.sender, input.value - value, input.id);
            utxos[change.id] = change;
        }

        delete utxos[input.id];

        emit Output(input.id, spent.owner, spent.value, spent.id);

        if (change.exists) {
            emit Output(input.id, change.owner, change.value, change.id);
        }
    }

    function depositMultisig(bytes32 idA, bytes32 idB, address addrA, address addrB, bytes sigA, bytes sigB) public {
        UTXO memory inputA = utxos[idA];
        UTXO memory inputB = utxos[idB];

        require(uint(addrA) < uint(addrB));
        require(inputA.exists || inputB.exists);

        if (inputA.exists) {
            require(inputA.owner == addrA);
        } else {
            require(idA == ZERO_BYTES32);
        }

        if (inputB.exists) {
            require(inputB.owner == addrB);
        } else {
            require(idB == ZERO_BYTES32);
        }

        bytes32 hash = hashMultisig(idA, idB, addrA, addrB);
        require(verify(hash, addrA, sigA));
        require(verify(hash, addrB, sigB));

        uint valueA = inputA.exists ? inputA.value : 0;
        uint valueB = inputB.exists ? inputB.value : 0;

        bytes32 id = genMultisigId(addrA, addrB, idA, idB);
        Multisig memory multi = Multisig(addrA, addrB, valueA, valueB, block.number, id, ZERO_BYTES32, 0, address(0), true);
        multisigs[multi.id] = multi;

        if (inputA.exists) {
            delete utxos[inputA.id];
        }

        if (inputB.exists) {
            delete utxos[inputB.id];
        }

        emit MultisigOutput(multi.signerA, multi.signerB, multi.valueA, multi.valueB, multi.id);
    }

    function spendMultisig(bytes32 id, bytes encumbranceBytes) public {
        Multisig memory multisig = multisigs[id];
        require(multisig.exists);
        require(msg.sender == multisig.signerA || msg.sender == multisig.signerB);
        require(multisig.hashLock == ZERO_BYTES32);

        Encumbrance memory encumbrance = deserializeEncumbrance(encumbranceBytes);
        require(encumbrance.lockTime > 0);
        require(multisig.valueA > 0);
        require(multisig.valueB > 0);
        require(multisig.valueA + multisig.valueB == encumbrance.valueA + encumbrance.valueB);

        bytes32 encumbranceHash = keccak256(encumbranceBytes.slice(0, ENCUMBRANCE_SIZE - 130));
        require(verify(encumbranceHash, multisig.signerA, encumbrance.sigA));
        require(verify(encumbranceHash, multisig.signerB, encumbrance.sigB));

        UTXO memory utxo;
        Multisig memory challenger;
        bytes32 challengerId;

        if (msg.sender == multisig.signerA) {
            utxo = makeUTXO(multisig.signerB, encumbrance.valueB, multisig.id);
            challengerId = genMultisigId(multisig.signerA, multisig.signerB, multisig.id, ZERO_BYTES32);
            challenger = Multisig(multisig.signerA, multisig.signerB, encumbrance.valueA, 0, block.number, challengerId, encumbrance.hashLock, encumbrance.lockTime, msg.sender, true);
        } else {
            utxo = makeUTXO(multisig.signerA, encumbrance.valueA, multisig.id);
            challengerId = genMultisigId(multisig.signerA, multisig.signerB, multisig.id, ZERO_BYTES32);
            challenger = Multisig(multisig.signerA, multisig.signerB, 0, encumbrance.valueB, block.number, challengerId, encumbrance.hashLock, encumbrance.lockTime, msg.sender, true);
        }

        utxos[utxo.id] = utxo;
        multisigs[challenger.id] = challenger;
        delete multisigs[multisig.id];
        
        emit Output(multisig.id, utxo.owner, utxo.value, utxo.id);
        emit MultisigOutput(challenger.signerA, challenger.signerB, challenger.valueA, challenger.valueB, challenger.id);
    }

    function settleMultisig(bytes32 id, bytes sigA, bytes sigB) public {
        Multisig memory multisig = multisigs[id];
        require(multisig.exists);
        require(msg.sender == multisig.signerA || msg.sender == multisig.signerB);

        BytesBuffer.Buffer memory buf = BytesBuffer.Buffer(new bytes(33), 0);
        buf.putByte(MULTISIG_SIGIL);
        buf.putBytes32(id);
        bytes32 hash = keccak256(buf.data);
        require(verify(hash, multisig.signerA, sigA));
        require(verify(hash, multisig.signerB, sigB));
        exitMultisig(multisig);
    }

    function challengeMultisig(bytes32 id, bytes32 preimage) public {
        Multisig memory multisig = multisigs[id];
        require(multisig.exists);
        require(msg.sender == multisig.signerA || msg.sender == multisig.signerB);
        require(multisig.hashLock != ZERO_BYTES32);
        require(multisig.lockedFor != msg.sender);

        address breacher;

        if (msg.sender == multisig.signerA) {
            breacher = multisig.signerB;
        } else {
            breacher = multisig.signerA;
        }

        bytes32 hash = keccak256(preimage);

        if (multisig.hashLock == hash) {
            uint total = multisig.valueA + multisig.valueB;
            UTXO memory breach = makeUTXO(msg.sender, total, multisig.id);
            utxos[breach.id] = breach;
            delete multisigs[multisig.id];
            emit Output(multisig.id, breach.owner, breach.value, breach.id);
            emit Breach(breacher, multisig.id);
            return;
        }

        revert();
    }

    function timeoutMultisig(bytes32 id) public {
        Multisig memory multisig = multisigs[id];
        require(multisig.exists);
        require(msg.sender == multisig.signerA || msg.sender == multisig.signerB);
        require(multisig.hashLock != ZERO_BYTES32);
        require(multisig.lockedFor == msg.sender);
        require(block.number > multisig.blockNum + multisig.lockTime);
        exitMultisig(multisig);
    }

    function hashMultisig(bytes32 idA, bytes32 idB, address addrA, address addrB) public returns (bytes32) {
        BytesBuffer.Buffer memory buf = BytesBuffer.Buffer(new bytes(104), 0);
        buf.putBytes32(idA);
        buf.putBytes32(idB);
        buf.putAddress(addrA);
        buf.putAddress(addrB);
        return keccak256(buf.data);
    }

    function exitMultisig(Multisig multisig) private {
        UTXO memory utxoA;
        UTXO memory utxoB;

        if (multisig.valueA > 0) {
            utxoA = makeUTXO(multisig.signerA, multisig.valueA, multisig.id);
            utxos[utxoA.id] = utxoA;
        }

        if (multisig.valueB > 0) {
            utxoB = makeUTXO(multisig.signerB, multisig.valueB, multisig.id);
            utxos[utxoB.id] = utxoB;
        }

        delete multisigs[multisig.id];

        if (utxoA.exists) {
            emit Output(multisig.id, utxoA.owner, utxoA.value, utxoA.id);
        }

        if (utxoB.exists) {
            emit Output(multisig.id, utxoB.owner, utxoB.value, utxoB.id);
        }
    }

    function genId(address owner, bytes32 inputId) private returns (bytes32) {
        return keccak256(block.number, owner, inputId);
    }

    function genMultisigId(address signerA, address signerB, bytes32 idA, bytes32 idB) private returns (bytes32) {
        return keccak256(idA, idB, signerA, signerB);
    }

    function verify(bytes32 data, address expected, bytes sig) returns (bool) {
        return ECRecovery.recover(ECRecovery.toEthSignedMessageHash(data), sig) == expected;
    }

    function deserializeEncumbrance(bytes encumbrance) private returns (Encumbrance) {
        require(encumbrance.length == ENCUMBRANCE_SIZE);

        uint i = 0;
        bytes32 id = encumbrance.slice32(i);
        i += 32;
        uint lockTime = encumbrance.toUint(i);
        i += 32;
        uint valueA = encumbrance.toUint(i);
        i += 32;
        uint valueB = encumbrance.toUint(i);
        i += 32;
        bytes32 hashLock = encumbrance.slice32(i);
        i += 32;
        bytes memory sigA = encumbrance.slice(i, 65);
        i += 65;
        bytes memory sigB = encumbrance.slice(i, 65);

        return Encumbrance(id, lockTime, valueA, valueB, hashLock, sigA, sigB);
    }

    function makeUTXO(address recipient, uint value, bytes32 inputId) private returns (UTXO) {
        bytes32 id = genId(recipient, inputId);
        return UTXO(recipient, value, block.number, id, true);
    }

    function hashEncumbrance(bytes encumbranceBytes) private returns (bytes32) {
        bytes memory toHash = encumbranceBytes.slice(0, encumbranceBytes.length - 130);
        return keccak256(toHash);
    }
}