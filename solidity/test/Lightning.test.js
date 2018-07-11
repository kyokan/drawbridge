const TestToken = artifacts.require('TestToken');
const LightningERC20 = artifacts.require('LightningERC20');
const abi = require('ethereumjs-abi');
const ethUtil = require('ethereumjs-util');
const BN = require('bn.js');

contract('LightningERC20', (accounts) => {
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

      it('should emit a Create event', async () => {
        assert.strictEqual(log.value.toNumber(), depositedTokens);
        assert.isNumber(log.blockNum.toNumber());
        assert.strictEqual(log.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
        assert.strictEqual(log.id, '0x9226c56484f590d8e79ebfe90766decaceac87e0066117868eed96c155a18fd1');
      });

      it('should store an Output', async () => {
        const id = log.id;
        const output = parseOutputStruct(await lightningContract.outputs.call(id));
        assert.strictEqual(output.value, depositedTokens);
        assert.isNumber(output.blockNum);
        assert.strictEqual(output.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
        assert.strictEqual(output.id, '0x9226c56484f590d8e79ebfe90766decaceac87e0066117868eed96c155a18fd1');
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
      let sent;

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
          await createPayableWitness(inputId, accounts[0], outputs),
          outputs
        );

        sent = res.logs[1].args;
        change = res.logs[2].args;
      });

      it('should emit the correct Create events', () => {
        assert.strictEqual(sent.value.toNumber(), sentAmount);
        assert.isNumber(sent.blockNum.toNumber());
        assert.strictEqual(sent.script, '0x01f17f52151ebef6c7334fad080c5704d77216b732');
        assert.strictEqual(sent.id, '0x40ad09dd47198ce4fe0a433bad89ff91e3f87ab46a134a4e53dccfcddf10316f');

        assert.strictEqual(change.value.toNumber(), mintedTokens - sentAmount);
        assert.isNumber(change.blockNum.toNumber());
        assert.strictEqual(change.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
        assert.strictEqual(change.id, '0x412152eeb2d6afc5819231fed25490a7034e54ad2cb810ff557396c46444801a');
      });

      it('should store the correct Output', async () => {
        const spentOut = parseOutputStruct(await lightningContract.outputs.call(sent.id));
        assert.strictEqual(spentOut.value, sentAmount);
        assert.isNumber(spentOut.blockNum);
        assert.strictEqual(spentOut.script, '0x01f17f52151ebef6c7334fad080c5704d77216b732');
        assert.strictEqual(spentOut.id, '0x40ad09dd47198ce4fe0a433bad89ff91e3f87ab46a134a4e53dccfcddf10316f');

        const changeOut = parseOutputStruct(await lightningContract.outputs.call(change.id));
        assert.strictEqual(changeOut.value, mintedTokens - sentAmount);
        assert.isNumber(changeOut.blockNum);
        assert.strictEqual(changeOut.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
        assert.strictEqual(changeOut.id, '0x412152eeb2d6afc5819231fed25490a7034e54ad2cb810ff557396c46444801a');
      });
    });

    describe('on success without change', () => {
      let sent;

      before(async () => {
        let res = await lightningContract.deposit(mintedTokens);
        const inputId = res.logs[0].args.id;
        const output = createPayableOutput(mintedTokens, accounts[1]);
        res = await lightningContract.spend(
          await createPayableWitness(inputId, accounts[0], output),
          output
        );
        sent = res.logs[1].args;
      });

      it('should emit the correct Create event', () => {
        assert.strictEqual(sent.value.toNumber(), mintedTokens);
        assert.isNumber(sent.blockNum.toNumber());
        assert.strictEqual(sent.script, '0x01f17f52151ebef6c7334fad080c5704d77216b732');
        assert.strictEqual(sent.id, '0x05d1136581d83d0d6efad1ba8a43618959891cbd208a90cfdb0ab944c9696160');
      });

      it('should store the correct utxos', async () => {
        const spentOut = parseOutputStruct(await lightningContract.outputs.call(sent.id));
        assert.strictEqual(spentOut.value, mintedTokens);
        assert.isNumber(spentOut.blockNum);
        assert.strictEqual(spentOut.script, '0x01f17f52151ebef6c7334fad080c5704d77216b732');
        assert.strictEqual(spentOut.id, '0x05d1136581d83d0d6efad1ba8a43618959891cbd208a90cfdb0ab944c9696160');
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

    describe('funding a two-input multisig', () => {
      let sent;

      let change;

      let multi;

      before(async () => {
        let res = await lightningContract.deposit(mintedTokens);
        const inputId = res.logs[0].args.id;
        const outputs = concatOutputs(
          createPayableOutput(sentAmount, accounts[1]),
          createPayableOutput(mintedTokens - sentAmount, accounts[0]),
        );
        res = await lightningContract.spend(
          await createPayableWitness(inputId, accounts[0], outputs),
          outputs
        );

        sent = res.logs[1].args;
        change = res.logs[2].args;

        const multisig = createMultisigOutput(mintedTokens, accounts[0], accounts[1]);
        const inputA = await createPayableWitness(sent.id, accounts[1], multisig);
        const inputB = await createPayableWitness(change.id, accounts[0], multisig);
        const inputs = concatOutputs(inputA, inputB);

        res = await lightningContract.spend(
          inputs,
          multisig
        );

        multi = res.logs[2].args;
      });

      it('should emit the correct Create event', () => {
        assert.strictEqual(multi.value.toNumber(), mintedTokens);
        assert.isNumber(multi.blockNum.toNumber());
        assert.strictEqual(multi.script, `0x02${strip0x(accounts[0])}${strip0x(accounts[1])}`);
        assert.strictEqual(multi.id, '0xc027addfdcc47e5674ffd32bfc735686082113b884c6eb14ce4bfca745fcde31');
      });

      it('should store the correct Output', async () => {
        const multiOut = parseOutputStruct(await lightningContract.outputs.call(multi.id));
        assert.strictEqual(multiOut.value, mintedTokens);
        assert.isNumber(multiOut.blockNum);
        assert.strictEqual(multiOut.script, `0x02${strip0x(accounts[0])}${strip0x(accounts[1])}`);
        assert.strictEqual(multiOut.id, '0xc027addfdcc47e5674ffd32bfc735686082113b884c6eb14ce4bfca745fcde31');
      });

      it('should revert if only one party signs', async () => {
        const output = createPayableOutput(mintedTokens, accounts[0]);
        const idBuf = Buffer.from(strip0x(multi.id), 'hex');
        const lenBuf = new BN(131).toArrayLike(Buffer, 'be', 16);
        const dataBuf = Buffer.from('00', 'hex');

        const sigBuf = Buffer.concat([
          idBuf,
          dataBuf,
          Buffer.from(strip0x(output), 'hex')
        ]);

        const hash = ethUtil.keccak256(sigBuf);

        const sig = await web3.eth.sign(accounts[0], '0x' + hash.toString('hex'));

        const inputBufZeroSig = Buffer.concat([
          idBuf,
          lenBuf,
          dataBuf,
          Buffer.from(strip0x(sig), 'hex'),
          Buffer.alloc(65)
        ]);

        await assertThrows(() => lightningContract.spend('0x' + inputBufZeroSig.toString('hex'), output));

        const inputBufSingleSig = Buffer.from([
          idBuf,
          new BN(66).toArrayLike(Buffer, 'be', 16),
          dataBuf,
          Buffer.from(strip0x(sig), 'hex'),
        ]);

        await assertThrows(() => lightningContract.spend('0x' + inputBufSingleSig.toString('hex'), output));
      });

      it('should allow spends if both parties sign', async () => {
        const changeOutput = createPayableOutput(mintedTokens - 100, accounts[0]);
        const spendOutput = createPayableOutput(100, accounts[1]);
        const outputs = concatOutputs(changeOutput, spendOutput);
        const idBuf = Buffer.from(strip0x(multi.id), 'hex');
        const lenBuf = new BN(131).toArrayLike(Buffer, 'be', 16);
        const dataBuf = Buffer.from('00', 'hex');

        const sigBuf = Buffer.concat([
          idBuf,
          dataBuf,
          Buffer.from(strip0x(outputs), 'hex')
        ]);

        const hash = ethUtil.keccak256(sigBuf);
        const sigA = await web3.eth.sign(accounts[0], '0x' + hash.toString('hex'));
        const sigB = await web3.eth.sign(accounts[1], '0x' + hash.toString('hex'));

        const inputBuf = Buffer.concat([
          idBuf,
          lenBuf,
          dataBuf,
          Buffer.from(strip0x(sigA), 'hex'),
          Buffer.from(strip0x(sigB), 'hex')
        ]);

        const res = await lightningContract.spend('0x' + inputBuf.toString('hex'), outputs);
        assert.strictEqual(res.logs[1].event, 'Create');
        assert.strictEqual(res.logs[2].event, 'Create');
        assert.strictEqual(res.logs[1].args.value.toNumber(), mintedTokens - 100);
        assert.strictEqual(res.logs[2].args.value.toNumber(), 100);
      });
    });

    describe('funding a one-input multisig', () => {
      let sent;

      let multi;

      before(async () => {
        let res = await lightningContract.deposit(mintedTokens);
        sent = res.logs[0].args;

        const multisig = createMultisigOutput(mintedTokens, accounts[0], accounts[1]);
        const inputA = await createPayableWitness(sent.id, accounts[0], multisig);

        res = await lightningContract.spend(
          inputA,
          multisig
        );

        multi = res.logs[1].args;
      });

      it('should emit the correct Create event', () => {
        assert.strictEqual(multi.value.toNumber(), mintedTokens);
        assert.isNumber(multi.blockNum.toNumber());
        assert.strictEqual(multi.script, `0x02${strip0x(accounts[0])}${strip0x(accounts[1])}`);
        assert.strictEqual(multi.id, '0xcb7aa776f33d1d417c2436438ca334da6f70bcf208b7492f1dc5c9433aafb39e');
      });
    });

    describe('to local commitments', () => {
      let commitment;

      beforeEach(async () => {
        let res = await lightningContract.deposit(mintedTokens);
        const inputId = res.logs[0].args.id;

        const output = createLocalCommitOutput(mintedTokens, 10, accounts[0], accounts[1]);

        res = await lightningContract.spend(
          await createPayableWitness(inputId, accounts[0], output),
          output
        );

        commitment = res.logs[1].args;
      });

      it('should not be spendable by the delayed key for locktime blocks', async () => {
        const idBuf = Buffer.from(strip0x(commitment.id), 'hex');
        const lenBuf = new BN(66).toArrayLike(Buffer, 'be', 16);
        const dataBuf = Buffer.from('00', 'hex');
        const output = createPayableOutput(mintedTokens, accounts[0]);

        const sigBuf = Buffer.concat([
          idBuf,
          dataBuf,
          Buffer.from(strip0x(output), 'hex')
        ]);

        const hash = ethUtil.keccak256(sigBuf);
        const sig = await web3.eth.sign(accounts[0], '0x' + hash.toString('hex'));

        const inputBuf = Buffer.concat([
          idBuf,
          lenBuf,
          dataBuf,
          Buffer.from(strip0x(sig), 'hex')
        ]);

        await assertThrows(() => lightningContract.spend('0x' + inputBuf.toString('hex'), output));

        // skip 11 blocks
        for (let i = 0; i < 11; i++) {
          await tokenContract.mint(accounts[5], 100);
        }

        const res = await lightningContract.spend('0x' + inputBuf.toString('hex'), output);
        assert.strictEqual(res.logs[1].event, 'Create');
        assert.strictEqual(res.logs[1].args.value.toNumber(), mintedTokens);
      });

      it('should be spendable by the revocation key at any time', async () => {
        const idBuf = Buffer.from(strip0x(commitment.id), 'hex');
        const lenBuf = new BN(66).toArrayLike(Buffer, 'be', 16);
        const dataBuf = Buffer.from('01', 'hex');
        const output = createPayableOutput(mintedTokens, accounts[1]);

        const sigBuf = Buffer.concat([
          idBuf,
          dataBuf,
          Buffer.from(strip0x(output), 'hex')
        ]);

        const hash = ethUtil.keccak256(sigBuf);
        const sig = await web3.eth.sign(accounts[1], '0x' + hash.toString('hex'));

        const inputBuf = Buffer.concat([
          idBuf,
          lenBuf,
          dataBuf,
          Buffer.from(strip0x(sig), 'hex')
        ]);

        const res = await lightningContract.spend('0x' + inputBuf.toString('hex'), output);
        assert.strictEqual(res.logs[1].event, 'Create');
        assert.strictEqual(res.logs[1].args.value.toNumber(), mintedTokens);
      });
    });

    describe('offered HTLCs', () => {
      let sent;

      let htlc;

      beforeEach(async () => {
        let res = await lightningContract.deposit(mintedTokens);
        sent = res.logs[0].args;

        const multisig = createMultisigOutput(mintedTokens, accounts[0], accounts[1]);
        const inputA = await createPayableWitness(sent.id, accounts[0], multisig);

        res = await lightningContract.spend(
          inputA,
          multisig
        );

        const multi = res.logs[1].args;
        const htlcOutput = createHtlcOutput(mintedTokens, 10, accounts[1], accounts[0], '0x0000000000000000000000000000000000000000000000000000000000000000');
        const multisigWitness = await createMultisigWitness(multi.id, accounts[0], accounts[1], htlcOutput);
        res = await lightningContract.spend(
          multisigWitness,
          htlcOutput,
        );
        htlc = res.logs[1].args;
      });

      it('should emit the correct Create event', () => {
        assert.strictEqual(htlc.value.toNumber(), mintedTokens);
        assert.isNumber(htlc.blockNum.toNumber());
        assert.strictEqual(htlc.script, `0x04000000000000000000000000000000000000000000000000000000000000000af17f52151ebef6c7334fad080c5704d77216b732627306090abab3a6e1400e9345bc60c78a8bef5766687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f2925`);
        assert.strictEqual(htlc.id, '0xee1f1e4593da8661e3be6e2c6a85c79bfcab6d10767bd99ce087b016e7c8adae');
      });

      it('should be immediately spendable by providing the preimage', async () => {
        const htlcWitness = await createHtlcWitness(
          htlc.id,
          accounts[1],
          '0x0000000000000000000000000000000000000000000000000000000000000000',
        );

        const output = createPayableOutput(mintedTokens, accounts[0]);
        const res = await lightningContract.spend(
          htlcWitness,
          output,
          {
            from: accounts[1],
          }
        );
        const payable = res.logs[1].args;
        assert.strictEqual(payable.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
      });

      it('should revert if timeout is called before the locktime', async () => {
        const htlcWitness = await createHtlcWitness(
          htlc.id,
          accounts[0],
          '',
        );

        const output = createPayableOutput(mintedTokens, accounts[0]);
        await assertThrows(() => lightningContract.spend(
          htlcWitness,
          output
        ));
      });

      it('should spend if timeout is called after the locktime', async () => {
        // skip 11 blocks
        for (let i = 0; i < 11; i++) {
          await tokenContract.mint(accounts[5], 100);
        }

        const htlcWitness = await createHtlcWitness(
          htlc.id,
          accounts[0],
          '',
        );
        const output = createPayableOutput(mintedTokens, accounts[0]);
        const res = await lightningContract.spend(
          htlcWitness,
          output
        );
        const payable = res.logs[1].args;
        assert.strictEqual(payable.script, '0x01627306090abab3a6e1400e9345bc60c78a8bef57');
      });
    });
  });

  describe('#withdraw', () => {
    const sentTokens = 1000;

    let change;

    let sent;

    beforeEach(async () => {
      await tokenContract.mint(accounts[2], mintedTokens);
      await tokenContract.approve(lightningContract.address, mintedTokens, {
        from: accounts[2]
      });

      const dep = await lightningContract.deposit(mintedTokens, {
        from: accounts[2]
      });
      const outputs = concatOutputs(
        createPayableOutput(sentTokens, accounts[3]),
        createPayableOutput(mintedTokens - sentTokens, accounts[2]),
      );
      const res = await lightningContract.spend(
        await createPayableWitness(dep.logs[0].args.id, accounts[2], outputs),
        outputs
      );
      sent = res.logs[1].args;
      change = res.logs[2].args;
    });

    it('should credit the token holder', async () => {
      await lightningContract.withdraw(await createPayableWitness(change.id, accounts[2], ZERO_BYTES), accounts[2]);
      await lightningContract.withdraw(await createPayableWitness(sent.id, accounts[3], ZERO_BYTES), accounts[3]);
      const changeBalance = await tokenContract.balanceOf.call(accounts[2]);
      const spentBalance = await tokenContract.balanceOf.call(accounts[3]);
      assert.strictEqual(changeBalance.toNumber(), 99000);
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
      lightningContract = await LightningERC20.new(tokenContract.address);

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
    value: indexed[0].toNumber(),
    blockNum: indexed[1].toNumber(),
    script: indexed[2],
    id: indexed[3],
  };
}

async function createPayableWitness(outputId, address, outputs) {
  const witnessBuf = Buffer.concat([
    Buffer.from(strip0x(outputId), 'hex'),
    Buffer.from('00', 'hex')
  ]);

  const hash = ethUtil.keccak256(Buffer.concat([
    witnessBuf,
    Buffer.from(strip0x(outputs), 'hex')
  ]));

  const sig = await web3.eth.sign(address, '0x' + hash.toString('hex'));

  return '0x' + Buffer.concat([
    Buffer.from(strip0x(outputId), 'hex'),
    new BN(66).toArrayLike(Buffer, 'be', 16),
    Buffer.from('00', 'hex'),
    Buffer.from(strip0x(sig), 'hex')
  ]).toString('hex')
}

async function createMultisigWitness(outputId, addressA, addressB, outputs) {
  const hash = ethUtil.keccak256(Buffer.concat([
    Buffer.from(strip0x(outputId), 'hex'),
    Buffer.from('00', 'hex'),
    Buffer.from(strip0x(outputs), 'hex')
  ]));

  const sigA = await web3.eth.sign(addressA, '0x' + hash.toString('hex'));
  const sigB = await web3.eth.sign(addressB, '0x' + hash.toString('hex'));

  return '0x' + Buffer.concat([
    Buffer.from(strip0x(outputId), 'hex'),
    new BN(131).toArrayLike(Buffer, 'be', 16),
    Buffer.from('00', 'hex'),
    Buffer.from(strip0x(sigA), 'hex'),
    Buffer.from(strip0x(sigB), 'hex'),
  ]).toString('hex');
}

async function createHtlcWitness(outputId, address, preimage) {
  let sig = Buffer.from('');
  if (!preimage) {
    const hash = ethUtil.keccak256(Buffer.concat([
      Buffer.from(strip0x(outputId), 'hex'),
      Buffer.from('01', 'hex'),
    ]));

    sig = Buffer.from(strip0x(await web3.eth.sign(address, '0x' + hash.toString('hex'))), 'hex');
  }

  return '0x' + Buffer.concat([
    Buffer.from(strip0x(outputId), 'hex'),
    new BN(preimage ? 33 : 66).toArrayLike(Buffer, 'be', 16),
    Buffer.from(preimage ? '00' : '01', 'hex'),
    Buffer.from(strip0x(preimage), 'hex'),
    sig
  ]).toString('hex')
}

function createPayableOutput(value, recipient) {
  const buf = abi.rawEncode([
    'uint',
  ], [
    value,
  ]);

  return '0x' + Buffer.concat([
    buf,
    Buffer.from('01', 'hex'),
    Buffer.from(strip0x(recipient), 'hex')
  ]).toString('hex');
}

function createMultisigOutput(value, recipA, recipB) {
  const buf = abi.rawEncode(['uint'], [value]);

  return '0x' + Buffer.concat([
    buf,
    Buffer.from('02', 'hex'),
    Buffer.from(strip0x(recipA), 'hex'),
    Buffer.from(strip0x(recipB), 'hex')
  ]).toString('hex')
}

function createLocalCommitOutput(value, delay, delayedAddress, revocationAddress) {
  const buf = abi.rawEncode(['uint'], [value]);
  const delayBuf = abi.rawEncode(['uint'], [delay]);

  return '0x' + Buffer.concat([
    buf,
    Buffer.from('03', 'hex'),
    delayBuf,
    Buffer.from(strip0x(delayedAddress), 'hex'),
    Buffer.from(strip0x(revocationAddress), 'hex'),
  ]).toString('hex');
}

function createHtlcOutput(value, delay, redemptionAddress, timeoutAddress, preimage) {
  const valueBuf = abi.rawEncode(['uint'], [value]);
  const delayBuf = abi.rawEncode(['uint'], [delay]);
  const hash = ethUtil.sha256(preimage);
  return '0x' + Buffer.concat([
    valueBuf,
    Buffer.from('04', 'hex'),
    delayBuf,
    Buffer.from(strip0x(redemptionAddress), 'hex'),
    Buffer.from(strip0x(timeoutAddress), 'hex'),
    hash,
  ]).toString('hex');
}

function concatOutputs(...outputs) {
  return '0x' + Buffer.concat(outputs.map((o) => Buffer.from(strip0x(o), 'hex'))).toString('hex')
}

function strip0x(hex) {
  return hex.replace('0x', '');
}