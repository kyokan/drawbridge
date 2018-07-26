const TestToken = artifacts.require('TestToken');
const Lightning = artifacts.require('Lightning');
const abi = require('ethereumjs-abi');
const ethUtil = require('ethereumjs-util');
const BN = require('bn.js');

contract('Lightning', (accounts) => {
  const ZERO_32 = '0x0000000000000000000000000000000000000000000000000000000000000000';

  const ZERO_BYTES = '0x0';

  const mintedTokens = 100000;

  let tokenContract;

  let lightningContract;

  redeployContract();

  beforeEach(async () => {
    await tokenContract.mint(accounts[0], mintedTokens);
    await tokenContract.approve(lightningContract.address, mintedTokens);
  });

  describe('#deposit', () => {
    const depositedTokens = mintedTokens - 1000;

    describe('on success', () => {
      let res;
      let log;

      before(async () => {
        res = await lightningContract.deposit(depositedTokens);
        log = res.logs[0].args;
      });

      it('should emit a NewOutput event', async () => {
        assert.strictEqual(log.lockTime.toNumber(), 0);
        assert.strictEqual(log.value.toNumber(), depositedTokens);
        assert.isNumber(log.blockNum.toNumber());
        assert.strictEqual(log.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
        assert.strictEqual(log.id.toNumber(), 1);
      });

      it('should store an Output', async () => {
        const id = log.id;
        const output = parseOutputStruct(await lightningContract.outputs.call(id));
        assert.strictEqual(output.lockTime, 0);
        assert.strictEqual(output.value, depositedTokens);
        assert.isNumber(output.blockNum);
        assert.strictEqual(output.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
        assert.strictEqual(output.id, 1);
      });
    });

    it('should revert if value is more than the allowance balance', async () => {
      await assertThrows(() => lightningContract.deposit(mintedTokens + 1));
    });

    it('should revert if value is less than zero', async () => {
      await assertThrows(() => lightningContract.deposit(-1));
    });

    it('should revert if the sender does not have a balance', async () => {
      await assertThrows(() => lightningContract.deposit(depositedTokens, {
        from: accounts[1]
      }));
    });
  });

  describe('#spend', () => {
    const sentAmount = 1000;

    describe('on success with change', () => {
      let spent;

      let change;

      let inputId;

      before(async () => {
        let res = await lightningContract.deposit(mintedTokens);
        inputId = res.logs[0].args.id;
        const outputs = concatOutputs(
          createPayableOutput(sentAmount, accounts[1]),
          createPayableOutput(mintedTokens - sentAmount, accounts[0]),
        );
        res = await lightningContract.spend(
          await createPayableInput(inputId, accounts[0], outputs),
          outputs
        );

        spent = res.logs[0].args;
        change = res.logs[1].args;
      });

      it('should emit the correct NewOutput events', () => {
        assert.strictEqual(spent.lockTime.toNumber(), 0);
        assert.strictEqual(spent.value.toNumber(), sentAmount);
        assert.isNumber(spent.blockNum.toNumber());
        assert.strictEqual(spent.script, '0x01f17f52151ebef6c7334fad080c5704d77216b732');
        assert.strictEqual(spent.id.toNumber(), 3);

        assert.strictEqual(change.lockTime.toNumber(), 0);
        assert.strictEqual(change.value.toNumber(), mintedTokens - sentAmount);
        assert.isNumber(change.blockNum.toNumber());
        assert.strictEqual(change.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
        assert.strictEqual(change.id.toNumber(), 4);
      });

      it('should store the correct utxos', async () => {
        const spentOut = parseOutputStruct(await lightningContract.outputs.call(spent.id));
        assert.strictEqual(spentOut.lockTime, 0);
        assert.strictEqual(spentOut.value, sentAmount);
        assert.isNumber(spentOut.blockNum);
        assert.strictEqual(spentOut.script, '0x01f17f52151ebef6c7334fad080c5704d77216b732');
        assert.strictEqual(spentOut.id, 3);

        const changeOut = parseOutputStruct(await lightningContract.outputs.call(change.id));
        assert.strictEqual(changeOut.lockTime, 0);
        assert.strictEqual(changeOut.value, mintedTokens - sentAmount);
        assert.isNumber(changeOut.blockNum);
        assert.strictEqual(changeOut.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
        assert.strictEqual(changeOut.id, 4);
      });
    });

    describe('on success without change', () => {
      let spent;

      before(async () => {
        let res = await lightningContract.deposit(mintedTokens);
        const inputId = res.logs[0].args.id;
        const output = createPayableOutput(mintedTokens, accounts[1]);
        res = await lightningContract.spend(
          await createPayableInput(inputId, accounts[0], output),
          output
        );
        spent = res.logs[0].args;
      });

      it('should emit the correct NewOutput event', () => {
        assert.strictEqual(spent.lockTime.toNumber(), 0);
        assert.strictEqual(spent.value.toNumber(), mintedTokens);
        assert.isNumber(spent.blockNum.toNumber());
        assert.strictEqual(spent.script, '0x01f17f52151ebef6c7334fad080c5704d77216b732');
        assert.strictEqual(spent.id.toNumber(), 6);
      });

      it('should store the correct utxos', async () => {
        const spentOut = parseOutputStruct(await lightningContract.outputs.call(spent.id));
        assert.strictEqual(spentOut.lockTime, 0);
        assert.strictEqual(spentOut.value, mintedTokens);
        assert.isNumber(spentOut.blockNum);
        assert.strictEqual(spentOut.script, '0x01f17f52151ebef6c7334fad080c5704d77216b732');
        assert.strictEqual(spentOut.id, 6);
      });
    });

    describe('on failures', () => {
      let inputId;

      before(async () => {
        let res = await lightningContract.deposit(mintedTokens);
        inputId = res.logs[0].args.id;
      });

      it('should revert if spending non-owned UTXOs', async () => {
        await assertThrows(() => lightningContract.spend(inputId, accounts[1], mintedTokens, ZERO_BYTES, {
          from: accounts[1]
        }));
      });

      it('should revert if the value is negative', async () => {
        await assertThrows(() => lightningContract.spend(inputId, accounts[1], -1, ZERO_BYTES));
      });

      it('should revert if the UTXO does not exist', async () => {
        await assertThrows(() => lightningContract.spend(ZERO_32, accounts[1], mintedTokens, ZERO_BYTES));
      });

      it('should revert if the value is higher than the UTXO', async () => {
        await assertThrows(() => lightningContract.spend(inputId, accounts[1], mintedTokens + 1, ZERO_BYTES));
      });
    });
  });

  describe('#withdraw', () => {
    const sentTokens = 1000;

    let change;

    let spent;

    beforeEach(async () => {
      const dep = await lightningContract.deposit(mintedTokens);
      const outputs = concatOutputs(
        createPayableOutput(sentTokens, accounts[1]),
        createPayableOutput(mintedTokens - sentTokens, accounts[0]),
      );
      const res = await lightningContract.spend(
        await createPayableInput(dep.logs[0].args.id, accounts[0], outputs),
        outputs
      );
      spent = res.logs[0].args;
      change = res.logs[1].args;
    });

    it('should credit the token holder', async () => {
      await lightningContract.withdraw(await createPayableInput(change.id, accounts[0], ZERO_BYTES), accounts[0]);
      await lightningContract.withdraw(await createPayableInput(spent.id, accounts[1], ZERO_BYTES), accounts[1]);
      const changeBalance = await tokenContract.balanceOf.call(accounts[0]);
      const spentBalance = await tokenContract.balanceOf.call(accounts[1]);
      assert.strictEqual(changeBalance.toNumber(), 1100000);
      assert.strictEqual(spentBalance.toNumber(), 1000);
    });

    it('should revert if the UTXO does not exist', async () => {
      await assertThrows(() => lightningContract.withdraw(ZERO_32, ZERO_BYTES))
    });

    it('should revert if spending another person\'s UTXOs', async () => {
      await assertThrows(() => lightningContract.withdraw(change.id, ZERO_BYTES, {
        from: accounts[1]
      }));
    });
  });

  function redeployContract() {
    before(async () => {
      tokenContract = await TestToken.new();
      lightningContract = await Lightning.new(tokenContract.address);

      await tokenContract.mint(accounts[0], mintedTokens);
      await tokenContract.approve(lightningContract.address, mintedTokens);
    });
  }
});

async function assertThrows(func) {
  try {
    await func();
  } catch (e) {
    return;
  }

  throw new Error('Expected error.');
}

function parseOutputStruct(indexed) {
  return {
    lockTime: indexed[0].toNumber(),
    value: indexed[1].toNumber(),
    blockNum: indexed[2].toNumber(),
    script: indexed[3],
    id: indexed[4].toNumber(),
  };
}

async function createPayableInput(outputId, address, outputs) {
  const witnessBuf = Buffer.concat([
    new BN(outputId.toString()).toArrayLike(Buffer, 'be', 32),
    new BN(20).toArrayLike(Buffer, 'be', 16),
    Buffer.from(strip0x(address), 'hex')
  ]);

  const hash = ethUtil.keccak256(Buffer.concat([
    witnessBuf,
    Buffer.from(strip0x(outputs), 'hex')
  ]));

  const sig = await web3.eth.sign(address, '0x' + hash.toString('hex'));

  return '0x' + Buffer.concat([
    witnessBuf,
    Buffer.from(strip0x(sig), 'hex')
  ]).toString('hex')
}

function createPayableOutput(value, recipient) {
  const buf = abi.rawEncode([
    'uint',
    'uint',
  ], [
    0,
    value,
  ]);

  return '0x' + Buffer.concat([
    buf,
    Buffer.from('01', 'hex'),
    Buffer.from(strip0x(recipient), 'hex')
  ]).toString('hex');
}

function concatOutputs(...outputs) {
  return '0x' + Buffer.concat(outputs.map((o) => Buffer.from(strip0x(o), 'hex'))).toString('hex')
}

function strip0x(hex) {
  return hex.replace('0x', '');
}