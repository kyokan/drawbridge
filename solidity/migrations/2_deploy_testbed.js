var TestToken = artifacts.require('./TestToken.sol');
var LightningERC20 = artifacts.require('./LightningERC20.sol');

module.exports = function (deployer) {
  deployer.deploy(TestToken)
    .then((instance) => {
      console.log('  Minting 1 million tokens for account 0.');

      return instance.mint('0x627306090abab3a6e1400e9345bc60c78a8bef57', 1000000)
        .then(() => instance);
    })
    .then((instance) => deployer.deploy(LightningERC20, instance.address))
};
