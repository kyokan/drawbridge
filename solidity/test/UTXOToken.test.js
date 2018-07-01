const TestToken = artifacts.require('TestToken');
const UTXOToken = artifacts.require('UTXOToken');
const crypto = require('crypto');

const MULTISIG_SIGIL = '02';

contract('UTXOToken', (accounts) => {
  const ZERO_32 = '0x0000000000000000000000000000000000000000000000000000000000000000';

  const ZERO_ADDRESS = '0x0000000000000000000000000000000000000000';

  const mintedTokens = 100000;

  let tokenContract;

  let utxoContract;

  beforeEach(async () => {
    tokenContract = await TestToken.new();
    utxoContract = await UTXOToken.new(tokenContract.address);

    await tokenContract.mint(accounts[0], mintedTokens);
    await tokenContract.approve(utxoContract.address, mintedTokens);
  });

  describe('#deposit', () => {
    const depositedTokens = mintedTokens - 1000;

    it('should emit a Deposit event', async () => {
      const res = await utxoContract.deposit(depositedTokens);
      const log = res.logs[0].args;
      assert.strictEqual(log.owner, accounts[0]);
      assert.strictEqual(log.value.toNumber(), depositedTokens);
    });

    it('should store a UTXO', async () => {
      const res = await utxoContract.deposit(depositedTokens);
      const id = res.logs[0].args.id;
      const utxo = parseUtxoStruct(await utxoContract.utxos.call(id));

      assert.strictEqual(utxo.owner, accounts[0]);
      assert.strictEqual(utxo.value, depositedTokens);
      assert.isNumber(utxo.blockNum);
      assert.strictEqual(utxo.id, id);
    });

    it('should revert if value is more than the allowance balance', async () => {
      await assertThrows(() => utxoContract.deposit(mintedTokens + 1));
    });

    it('should revert if value is less than zero', async () => {
      await assertThrows(() => utxoContract.deposit(-1));
    });

    it('should revert if the sender does not have a balance', async () => {
      await assertThrows(() => utxoContract.deposit(depositedTokens, {
        from: accounts[1]
      }));
    });
  });

  describe('#spend', () => {
    const sentAmount = 1000;

    let inputId;

    beforeEach(async () => {
      const res = await utxoContract.deposit(mintedTokens);
      inputId = res.logs[0].args.id;
    });

    describe('on success with change', () => {
      let change;

      let spent;

      beforeEach(async () => {
        const res = await utxoContract.spend(inputId, accounts[1], sentAmount);
        spent = res.logs[0].args;
        change = res.logs[1].args;
      });

      it('should emit the correct Output events', () => {
        assert.strictEqual(spent.owner, accounts[1]);
        assert.strictEqual(spent.value.toNumber(), sentAmount);
        assert.isString(spent.id);
        assert.strictEqual(change.owner, accounts[0]);
        assert.strictEqual(change.value.toNumber(), mintedTokens - sentAmount);
        assert.isString(change.id);
      });

      it('should store the correct utxos', async () => {
        const changeUtxo = parseUtxoStruct(await utxoContract.utxos.call(change.id));

        assert.strictEqual(changeUtxo.owner, accounts[0]);
        assert.strictEqual(changeUtxo.value, mintedTokens - sentAmount);
        assert.isNumber(changeUtxo.blockNum);
        assert.strictEqual(changeUtxo.id, change.id);

        const spentUtxo = parseUtxoStruct(await utxoContract.utxos.call(spent.id));

        assert.strictEqual(spentUtxo.owner, accounts[1]);
        assert.strictEqual(spentUtxo.value, sentAmount);
        assert.isNumber(spentUtxo.blockNum);
        assert.strictEqual(spentUtxo.id, spent.id);
      });
    });

    describe('on success without change', () => {
      let spent;

      beforeEach(async () => {
        const res = await utxoContract.spend(inputId, accounts[1], mintedTokens);
        spent = res.logs[0].args;
      });

      it('should emit the correct Spend event', () => {
        assert.strictEqual(spent.owner, accounts[1]);
        assert.strictEqual(spent.value.toNumber(), mintedTokens);
        assert.isString(spent.id);
      });

      it('should store the correct utxos', async () => {
        const spentUtxo = parseUtxoStruct(await utxoContract.utxos.call(spent.id));

        assert.strictEqual(spentUtxo.owner, accounts[1]);
        assert.strictEqual(spentUtxo.value, mintedTokens);
        assert.isNumber(spentUtxo.blockNum);
        assert.strictEqual(spentUtxo.id, spent.id);
      });
    });

    it('should revert if spending non-owned UTXOs', async () => {
      await assertThrows(() => utxoContract.spend(inputId, accounts[1], mintedTokens, {
        from: accounts[1]
      }));
    });

    it('should revert if the value is negative', async () => {
      await assertThrows(() => utxoContract.spend(inputId, accounts[1], -1));
    });

    it('should revert if the UTXO does not exist', async () => {
      await assertThrows(() => utxoContract.spend(ZERO_32, accounts[1], mintedTokens));
    });

    it('should revert if the value is higher than the UTXO', async () => {
      await assertThrows(() => utxoContract.spend(inputId, accounts[1], mintedTokens + 1));
    });
  });

  describe('#withdraw', () => {
    const sentTokens = 1000;

    let change;

    let spent;

    beforeEach(async () => {
      const dep = await utxoContract.deposit(mintedTokens);
      const res = await utxoContract.spend(dep.logs[0].args.id, accounts[1], sentTokens);
      spent = res.logs[0].args;
      change = res.logs[1].args;
    });

    it('should credit the token holder', async () => {
      await utxoContract.withdraw(change.id);
      await utxoContract.withdraw(spent.id, {
        from: accounts[1],
      });
      const changeBalance = await tokenContract.balanceOf.call(accounts[0]);
      const spentBalance = await tokenContract.balanceOf.call(accounts[1]);
      assert.strictEqual(changeBalance.toNumber(), 99000);
      assert.strictEqual(spentBalance.toNumber(), 1000);
    });

    it('should revert if the UTXO does not exist', async () => {
      await assertThrows(() => utxoContract.withdraw(ZERO_32))
    });

    it('should revert if spending another person\'s UTXOs', async () => {
      await assertThrows(() => utxoContract.withdraw(change.id, {
        from: accounts[1]
      }));
    });
  });

  describe('#depositMultisig', () => {
    let inputA;

    let inputB;

    beforeEach(async () => {
      const dep = await utxoContract.deposit(mintedTokens);
      const res = await utxoContract.spend(dep.logs[0].args.id, accounts[1], mintedTokens - 1000);
      inputA = res.logs[1].args;
      inputB = res.logs[0].args;
    });

    it('should properly hash the inputs', async () => {
      const sha = hashMultisig(inputA, inputB);
      const res = await utxoContract.hashMultisig.call(inputA.id, inputB.id, accounts[0], accounts[1]);
      assert.strictEqual(sha, res);
    });

    describe('when both sides sign the message', () => {
      let log;

      beforeEach(async () => {
        const sha = hashMultisig(inputA, inputB);
        const sigA = await web3.eth.sign(accounts[0], sha);
        const sigB = await web3.eth.sign(accounts[1], sha);

        const res = await utxoContract.depositMultisig(inputA.id, inputB.id, accounts[0], accounts[1], sigA, sigB);
        log = res.logs[0].args;
      });

      it('should emit a MultisigOutput event', async () => {
        assert.strictEqual(log.signerA, accounts[0]);
        assert.strictEqual(log.signerB, accounts[1]);
        assert.strictEqual(log.valueA.toNumber(), 1000);
        assert.strictEqual(log.valueB.toNumber(), mintedTokens - 1000);
        assert.isString(log.id);
      });

      it('should store a Multisig', async () => {
        const multisig = await utxoContract.multisigs.call(log.id);

        assert.strictEqual(multisig[0], accounts[0]);
        assert.strictEqual(multisig[1], accounts[1]);
        assert.strictEqual(multisig[2].toNumber(), 1000);
        assert.strictEqual(multisig[3].toNumber(), mintedTokens - 1000);
        assert.isTrue(multisig[4].gt(0));
        assert.isString(multisig[5]);
        assert.strictEqual(multisig[6], ZERO_32);
        assert.strictEqual(multisig[7].toNumber(), 0);
        assert.strictEqual(multisig[8], ZERO_ADDRESS);
        assert.isTrue(multisig[9]);
      });

      it('should remove the input UTXOs', async () => {
        const a = await utxoContract.utxos.call(inputA.id);
        const b = await utxoContract.utxos.call(inputB.id);

        // index 4 is the exists boolean
        assert.isFalse(a[4]);
        assert.isFalse(b[4]);
      });
    });
  });

  describe('#spendMultisig', () => {
    let multisig;

    beforeEach(async () => {
      const dep = await utxoContract.deposit(mintedTokens);
      const spendRes = await utxoContract.spend(dep.logs[0].args.id, accounts[1], mintedTokens - 1000);
      const inputA = spendRes.logs[1].args;
      const inputB = spendRes.logs[0].args;

      const sha = hashMultisig(inputA, inputB);
      const sigA = await web3.eth.sign(accounts[0], sha);
      const sigB = await web3.eth.sign(accounts[1], sha);

      const multisigRes = await utxoContract.depositMultisig(inputA.id, inputB.id, accounts[0], accounts[1], sigA, sigB);
      multisig = multisigRes.logs[0].args;
    });

    describe('with valid signatures', () => {
      let lockTime;

      let valueA;

      let valueB;

      let spendLogs;

      beforeEach(async () => {
        lockTime = 1000;
        valueA = mintedTokens - 5000;
        valueB = 5000;

        const encumbranceData = makeEncumbranceData(multisig.id, lockTime, valueA, valueB, 1234);
        const sha = sha256(encumbranceData);
        const sigA = await web3.eth.sign(accounts[0], sha);
        const sigB = await web3.eth.sign(accounts[1], sha);
        const encumbrance = Buffer.concat([
          encumbranceData,
          Buffer.from(strip0x(sigA), 'hex'),
          Buffer.from(strip0x(sigB), 'hex'),
        ]);
        const spendRes = await utxoContract.spendMultisig(multisig.id, '0x' + encumbrance.toString('hex'));
        spendLogs = spendRes.logs;
      });

      it('should send funds immediately to the side that is not making the call', () => {
        const outputA = spendLogs[0].args;

        assert.strictEqual(outputA.owner, accounts[1]);
        assert.strictEqual(outputA.value.toNumber(), 5000);
        assert.isString(outputA.id);
      });

      it('should send funds to an encumbered multisig for the side that is making the call', () => {
        const multisigB = spendLogs[1].args;
        assert.strictEqual(multisigB.signerA, accounts[0]);
        assert.strictEqual(multisigB.signerB, accounts[1]);
        assert.strictEqual(multisigB.valueA.toNumber(), mintedTokens - 5000);
        assert.strictEqual(multisigB.valueB.toNumber(), 0);
        assert.isString(multisigB.id);
      });

      it('should delete the multisig', async () => {
        const stored = await utxoContract.multisigs.call(multisig.id);
        assert.isFalse(stored[9]);
      });
    });
  });

  describe('#settleMultisig', () => {
    let multisig;

    beforeEach(async () => {
      const dep = await utxoContract.deposit(mintedTokens);
      const spendRes = await utxoContract.spend(dep.logs[0].args.id, accounts[1], mintedTokens - 1000);
      const inputA = spendRes.logs[1].args;
      const inputB = spendRes.logs[0].args;

      const sha = hashMultisig(inputA, inputB);
      const sigA = await web3.eth.sign(accounts[0], sha);
      const sigB = await web3.eth.sign(accounts[1], sha);

      const multisigRes = await utxoContract.depositMultisig(inputA.id, inputB.id, accounts[0], accounts[1], sigA, sigB);
      multisig = multisigRes.logs[0].args;
    });

    describe('when both parties sign', () => {
      let logs;

      beforeEach(async () => {
        const buf = Buffer.alloc(33);
        buf.write(MULTISIG_SIGIL, 0, 1, 'hex');
        buf.write(strip0x(multisig.id), 1, 32, 'hex');

        const sha = sha256(buf);
        const sigA = await web3.eth.sign(accounts[0], sha);
        const sigB = await web3.eth.sign(accounts[1], sha);

        const multisigRes = await utxoContract.settleMultisig(multisig.id, sigA, sigB);
        logs = multisigRes.logs;
      });

      it('should disburse the appropriate value to each side', () => {
        const toA = logs[0].args;
        assert.strictEqual(toA.owner, accounts[0]);
        assert.strictEqual(toA.value.toNumber(), 1000);
        assert.isString(toA.id);

        const toB = logs[1].args;
        assert.strictEqual(toB.owner, accounts[1]);
        assert.strictEqual(toB.value.toNumber(), mintedTokens - 1000);
        assert.isString(toB.id);
      });
    });
  });

  describe('#challengeMultisig', () => {
    describe('when the multisig is hashlocked', () => {
      let multisigSpendLogs;

      beforeEach(async () => {
        const dep = await utxoContract.deposit(mintedTokens);
        const spendRes = await utxoContract.spend(dep.logs[0].args.id, accounts[1], mintedTokens - 1000);
        const inputA = spendRes.logs[1].args;
        const inputB = spendRes.logs[0].args;

        let sha = hashMultisig(inputA, inputB);
        let sigA = await web3.eth.sign(accounts[0], sha);
        let sigB = await web3.eth.sign(accounts[1], sha);

        const multisigRes = await utxoContract.depositMultisig(inputA.id, inputB.id, accounts[0], accounts[1], sigA, sigB);
        const multisig = multisigRes.logs[0].args;

        const encumbranceData = makeEncumbranceData(multisig.id, 1000, 1000, mintedTokens - 1000, 1234);
        sha = sha256(encumbranceData);
        sigA = await web3.eth.sign(accounts[0], sha);
        sigB = await web3.eth.sign(accounts[1], sha);

        const encumbrance = Buffer.concat([
          encumbranceData,
          Buffer.from(strip0x(sigA), 'hex'),
          Buffer.from(strip0x(sigB), 'hex'),
        ]);

        const multisigSpendRes = await utxoContract.spendMultisig(multisig.id, '0x' + encumbrance.toString('hex'));
        multisigSpendLogs = multisigSpendRes.logs;
      });

      it('should send all funds to the party presenting the hashlock', async () => {
        const preimage = numToUint256BE(1234);
        const res = await utxoContract.challengeMultisig(multisigSpendLogs[1].args.id, '0x' + preimage.toString('hex'), {
          from: accounts[1]
        });
        const logs = res.logs;

        const output = logs[0].args;
        assert.strictEqual(output.owner, accounts[1]);

        const breach = logs[1].args;
        assert.strictEqual(breach.breachingAddress, accounts[0]);
        assert.strictEqual(breach.multisigId, multisigSpendLogs[1].args.id);
      });
    });
  });

  describe('#timeoutMultisig', () => {
    describe('when the lockTime expires', () => {
      let multisigSpendLogs;

      beforeEach(async () => {
        const dep = await utxoContract.deposit(mintedTokens);
        const spendRes = await utxoContract.spend(dep.logs[0].args.id, accounts[1], mintedTokens - 1000);
        const inputA = spendRes.logs[1].args;
        const inputB = spendRes.logs[0].args;

        let sha = hashMultisig(inputA, inputB);
        let sigA = await web3.eth.sign(accounts[0], sha);
        let sigB = await web3.eth.sign(accounts[1], sha);

        const multisigRes = await utxoContract.depositMultisig(inputA.id, inputB.id, accounts[0], accounts[1], sigA, sigB);
        const multisig = multisigRes.logs[0].args;

        const encumbranceData = makeEncumbranceData(multisig.id, 1, 1000, mintedTokens - 1000, 1234);
        sha = sha256(encumbranceData);
        sigA = await web3.eth.sign(accounts[0], sha);
        sigB = await web3.eth.sign(accounts[1], sha);
        const encumbrance = Buffer.concat([
          encumbranceData,
          Buffer.from(strip0x(sigA), 'hex'),
          Buffer.from(strip0x(sigB), 'hex'),
        ]);

        const multisigSpendRes = await utxoContract.spendMultisig(multisig.id, '0x' + encumbrance.toString('hex'));
        multisigSpendLogs = multisigSpendRes.logs;
      });

      it('should disburse money to the appropriate person', async () => {
        for (let i = 0; i < 2; i++) {
          // wait 2 blocks
          await tokenContract.mint(accounts[0], mintedTokens);
        }

        const res = await utxoContract.timeoutMultisig(multisigSpendLogs[1].args.id);
        const logs = res.logs;

        assert.strictEqual(logs.length, 1);
        
        const utxo = logs[0].args;
        
        assert.strictEqual(utxo.owner, accounts[0]);
        assert.strictEqual(utxo.value.toNumber(), 1000);
        assert.isString(utxo.id);
      });
    });
  });

  function hashMultisig(inputA, inputB) {
    const buf = Buffer.alloc(104);
    buf.write(strip0x(inputA.id), 0, 32, 'hex');
    buf.write(strip0x(inputB.id), 32, 32, 'hex');
    buf.write(strip0x(accounts[0]), 64, 20, 'hex');
    buf.write(strip0x(accounts[1]), 84, 20, 'hex');

    console.log(buf.toString('hex'));
    console.log(sha256(buf));

    return sha256(buf);
  }
});

function makeEncumbranceData(multisigId, lockTime, valueA, valueB, hashLock) {
  const multisigBuf = Buffer.alloc(32);
  multisigBuf.write(strip0x(multisigId), 0, 32, 'hex');

  return Buffer.concat([
    multisigBuf,
    numToUint256BE(lockTime),
    numToUint256BE(valueA),
    numToUint256BE(valueB),
    sha256(numToUint256BE(hashLock), true),
  ])
}

function numToUint256BE(num) {
  const buf = Buffer.alloc(4);
  buf.writeUInt32BE(num, 0);
  return Buffer.concat([Buffer.alloc(28), buf]);
}

async function assertThrows(func) {
  try {
    await func();
  } catch (e) {
    return;
  }

  throw new Error('Expected error.');
}

function strip0x(hex) {
  return hex.replace('0x', '');
}

function parseUtxoStruct(indexed) {
  return {
    owner: indexed[0],
    value: indexed[1].toNumber(),
    blockNum: indexed[2].toNumber(),
    id: indexed[3],
  }
}

function sha256(buf, retBuf) {
  const hash = crypto.createHash('sha256');
  hash.update(buf);
  const digest = hash.digest();
  return retBuf ? digest : '0x' + digest.toString('hex');
}